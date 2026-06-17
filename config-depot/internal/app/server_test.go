package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadDecryptsAndExtractsConfigBundle(t *testing.T) {
	dir := testDataDir(t)
	key := mustDecodeBase64(t, "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=")
	writeFile(t, dir, "upload_secret", "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=\n")
	writeFile(t, dir, "upload_token", "token-123\n")

	server, err := NewServer(Settings{DataDir: dir})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	bundle := makeTarGzip(t, map[string]string{
		"clash":      "clash config",
		"quan-basic": "quan basic config",
	})
	encrypted, err := EncryptUploadBundle(bundle, key)
	if err != nil {
		t.Fatalf("encrypt bundle: %v", err)
	}

	request := multipartUploadRequest(t, "/upload", "token-123", encrypted)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if response.Body.String() != "发布成功\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
	assertFileContent(t, filepath.Join(dir, "config", "clash"), "clash config")
	assertFileContent(t, filepath.Join(dir, "config", "quan-basic"), "quan basic config")
}

func TestSubRedirectsWhenSecretDoesNotMatch(t *testing.T) {
	dir := testDataDir(t)
	writeFile(t, dir, "sub_secret", "secret\n")

	server, err := NewServer(Settings{
		DataDir:    dir,
		BackendURL: "https://backend.example/sub?",
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/sub?target=quan&url=wrong", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d", response.Code)
	}
	if location := response.Header().Get("Location"); location != "https://backend.example/sub?target=quan&url=wrong" {
		t.Fatalf("location = %q", location)
	}
}

func TestSubReturnsCachedConfigWhenSecretMatches(t *testing.T) {
	dir := testDataDir(t)
	writeFile(t, dir, "sub_secret", "secret\n")
	writeFile(t, filepath.Join(dir, "config"), "quan-basic", "cached config")

	server, err := NewServer(Settings{DataDir: dir})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/sub?target=quan&url=secret-basic", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if response.Body.String() != "cached config" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestSubFetchesBackendAndCachesWhenCacheMiss(t *testing.T) {
	dir := testDataDir(t)
	writeFile(t, dir, "sub_secret", "secret\n")
	writeFile(t, dir, "subscribe", "https://example.com/a\n\nhttps://example.com/b\n")

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("target") != "clash" {
			t.Fatalf("target = %q", query.Get("target"))
		}
		if query.Get("url") != "https://example.com/a|https://example.com/b" {
			t.Fatalf("url = %q", query.Get("url"))
		}
		if query.Get("config") != "https://github.com/AoEiuV020/SubConfig/raw/main/subconverter-basic.ini" {
			t.Fatalf("config = %q", query.Get("config"))
		}
		_, _ = io.WriteString(w, "generated config")
	}))
	defer backend.Close()

	server, err := NewServer(Settings{
		DataDir:    dir,
		BackendURL: backend.URL + "/sub?",
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/sub?url=secret-basic", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if response.Body.String() != "generated config" {
		t.Fatalf("body = %q", response.Body.String())
	}
	assertFileContent(t, filepath.Join(dir, "config", "clash-basic"), "generated config")
}

func TestPrivateDataFilesAreNotServed(t *testing.T) {
	dir := testDataDir(t)
	writeFile(t, dir, "subscribe", "https://example.com/a\n")
	writeFile(t, dir, "sub_secret", "secret\n")
	writeFile(t, dir, "upload_secret", "secret\n")
	writeFile(t, filepath.Join(dir, "config"), "clash", "cached config")

	server, err := NewServer(Settings{DataDir: dir})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	for _, path := range []string{"/subscribe", "/sub_secret", "/upload_secret", "/config/clash"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d", path, response.Code)
		}
	}
}

func multipartUploadRequest(t *testing.T, targetURL string, token string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("token", token); err != nil {
		t.Fatalf("write token field: %v", err)
	}
	fileWriter, err := writer.CreateFormFile("file", "config.tar.gz.aes")
	if err != nil {
		t.Fatalf("create file field: %v", err)
	}
	if _, err := fileWriter.Write(content); err != nil {
		t.Fatalf("write file field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, targetURL, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func makeTarGzip(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("write tar content: %v", err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return buffer.Bytes()
}

func writeFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create dir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write file %s: %v", filepath.Join(dir, name), err)
	}
}

func assertFileContent(t *testing.T, path string, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(content) != expected {
		t.Fatalf("%s = %q", path, content)
	}
}

func testDataDir(t *testing.T) string {
	t.Helper()

	base := os.Getenv("CONFIG_DEPOT_TEST_TMP")
	if base == "" {
		base = filepath.Join("..", "..", "..", "tmp", "config-depot-tests")
	}
	base, err := filepath.Abs(base)
	if err != nil {
		t.Fatalf("resolve test tmp dir: %v", err)
	}

	name := strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(t.Name())
	dir := filepath.Join(base, name)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("clean test dir %s: %v", dir, err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create test dir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("清理测试目录失败: %v", err)
		}
		_ = os.Remove(base)
	})
	return dir
}
