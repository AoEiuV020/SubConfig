package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const maxUploadBytes = 128 << 20

var safeNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]*$`)

type Server struct {
	settings resolvedSettings
	mux      *http.ServeMux
}

func NewServer(settings Settings) (*Server, error) {
	resolved := resolveSettings(settings)
	server := &Server{
		settings: resolved,
		mux:      http.NewServeMux(),
	}
	server.routes()
	return server, nil
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.mux.ServeHTTP(w, r)
}

func (server *Server) Address() string {
	return server.settings.Address()
}

func (server *Server) routes() {
	server.mux.HandleFunc("/healthz", server.handleHealth)
	server.mux.HandleFunc("/upload", server.handleUpload)
	server.mux.HandleFunc("/sub", server.handleSub)
}

func (server *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "ok\n")
}

func (server *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "上传表单无效", http.StatusBadRequest)
		return
	}

	token, err := readRequiredTrimmedFile(server.settings.uploadToken)
	if err != nil {
		http.Error(w, "读取上传令牌失败", http.StatusInternalServerError)
		return
	}
	if r.PostFormValue("token") != token {
		http.Error(w, "上传令牌错误", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "缺少上传文件", http.StatusBadRequest)
		return
	}
	defer file.Close()
	if header.Filename != "config.tar.gz.aes" {
		http.Error(w, "上传文件名错误", http.StatusBadRequest)
		return
	}

	raw, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "读取上传文件失败", http.StatusBadRequest)
		return
	}

	keyText, err := readRequiredTrimmedFile(server.settings.uploadSecret)
	if err != nil {
		http.Error(w, "读取上传密钥失败", http.StatusInternalServerError)
		return
	}
	key, err := DecodeBase64Secret(keyText)
	if err != nil {
		http.Error(w, "上传密钥格式无效", http.StatusInternalServerError)
		return
	}
	bundle, err := DecryptUploadBundle(raw, key)
	if err != nil {
		http.Error(w, "解密上传文件失败", http.StatusBadRequest)
		return
	}
	if err := ExtractConfigBundle(bundle, server.settings.configDir); err != nil {
		http.Error(w, "解包配置失败", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "发布成功\n")
}

func (server *Server) handleSub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	secret, err := readRequiredTrimmedFile(server.settings.subSecret)
	if err != nil {
		http.Error(w, "读取订阅密钥失败", http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()
	secretURL := query.Get("url")
	if !strings.HasPrefix(secretURL, secret) {
		http.Redirect(w, r, appendRawQuery(server.settings.backendURL, r.URL.RawQuery), http.StatusMovedPermanently)
		return
	}

	suffix := strings.TrimPrefix(secretURL, secret)
	target := firstNonEmpty(query.Get("target"), "clash")
	ver := query.Get("ver")
	if err := validateSafePart("target", target); err != nil {
		http.Error(w, "target 参数无效", http.StatusBadRequest)
		return
	}
	if err := validateSafePart("ver", ver); err != nil {
		http.Error(w, "ver 参数无效", http.StatusBadRequest)
		return
	}
	if err := validateSafePart("url 后缀", suffix); err != nil {
		http.Error(w, "url 后缀无效", http.StatusBadRequest)
		return
	}

	cachePath := filepath.Join(server.settings.configDir, cacheFileName(target, ver, suffix))
	if shouldReadCache(query) {
		content, err := os.ReadFile(cachePath)
		if err == nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write(content)
			return
		}
		if !errors.Is(err, os.ErrNotExist) {
			http.Error(w, "读取缓存失败", http.StatusInternalServerError)
			return
		}
	}

	subscribeLines, err := readSubscribeLines(server.settings.subscribe)
	if err != nil {
		http.Error(w, "读取订阅文件失败", http.StatusInternalServerError)
		return
	}
	backendURL, err := server.buildBackendURL(target, ver, suffix, subscribeLines)
	if err != nil {
		http.Error(w, "构造后端请求失败", http.StatusInternalServerError)
		return
	}

	response, err := server.settings.httpClient.Get(backendURL)
	if err != nil {
		http.Error(w, "请求订阅转换后端失败", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "读取订阅转换结果失败", http.StatusBadGateway)
		return
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		http.Error(w, string(content), http.StatusBadGateway)
		return
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		http.Error(w, "创建缓存目录失败", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(cachePath, content, 0o644); err != nil {
		http.Error(w, "写入缓存失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}

func (server *Server) buildBackendURL(target string, ver string, suffix string, subscribeLines []string) (string, error) {
	backendURL, err := url.Parse(server.settings.backendURL)
	if err != nil {
		return "", err
	}

	values := backendURL.Query()
	values.Set("emoji", "true")
	values.Set("list", "false")
	values.Set("udp", "true")
	values.Set("tfo", "false")
	values.Set("scv", "false")
	values.Set("fdn", "false")
	values.Set("sort", "false")
	values.Set("new_name", "true")
	values.Set("target", target)
	values.Set("url", strings.Join(subscribeLines, "|"))
	values.Set("config", configURLForSuffix(server.settings.remoteConfigURL, suffix))
	if ver != "" {
		values.Set("ver", ver)
	}
	backendURL.RawQuery = values.Encode()
	return backendURL.String(), nil
}

func readTrimmedFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func readRequiredTrimmedFile(path string) (string, error) {
	value, err := readTrimmedFile(path)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", errors.New("required file is empty")
	}
	return value, nil
}

func readSubscribeLines(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	subscribeLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			subscribeLines = append(subscribeLines, line)
		}
	}
	return subscribeLines, nil
}

func appendRawQuery(baseURL string, rawQuery string) string {
	if rawQuery == "" {
		return strings.TrimSuffix(baseURL, "?")
	}
	if strings.HasSuffix(baseURL, "?") || strings.HasSuffix(baseURL, "&") {
		return baseURL + rawQuery
	}
	if strings.Contains(baseURL, "?") {
		return baseURL + "&" + rawQuery
	}
	return baseURL + "?" + rawQuery
}

func shouldReadCache(query url.Values) bool {
	cache := query.Get("cache")
	return cache == "" || cache == "true"
}

func cacheFileName(target string, ver string, suffix string) string {
	if ver != "" {
		return target + "-" + ver + suffix
	}
	return target + suffix
}

func configURLForSuffix(configURL string, suffix string) string {
	if suffix == "" || !strings.HasSuffix(configURL, ".ini") {
		return configURL
	}
	return strings.TrimSuffix(configURL, ".ini") + suffix + ".ini"
}

func validateSafePart(name string, value string) error {
	if !safeNamePattern.MatchString(value) {
		return fmt.Errorf("%s contains unsafe characters", name)
	}
	return nil
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
}
