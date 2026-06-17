package app

import (
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Settings struct {
	Address         string
	DataDir         string
	ConfigDir       string
	UploadSecret    string
	UploadToken     string
	SubSecret       string
	Subscribe       string
	BackendURL      string
	RemoteConfigURL string
	HTTPClient      *http.Client
}

func SettingsFromEnv() Settings {
	dataDir := getenv(EnvDataDir, DefaultDataDir)
	return Settings{
		Address:         getenv(EnvAddress, DefaultAddress),
		DataDir:         dataDir,
		ConfigDir:       getenv(EnvConfigDir, filepath.Join(dataDir, DefaultConfigDirName)),
		UploadSecret:    getenv(EnvUploadSecret, filepath.Join(dataDir, DefaultUploadSecret)),
		UploadToken:     getenv(EnvUploadToken, filepath.Join(dataDir, DefaultUploadToken)),
		SubSecret:       getenv(EnvSubSecret, filepath.Join(dataDir, DefaultSubSecret)),
		Subscribe:       getenv(EnvSubscribe, filepath.Join(dataDir, DefaultSubscribe)),
		BackendURL:      getenv(EnvBackendURL, DefaultBackendURL),
		RemoteConfigURL: getenv(EnvRemoteConfigURL, DefaultRemoteConfigURL),
	}
}

type resolvedSettings struct {
	address         string
	dataDir         string
	configDir       string
	uploadSecret    string
	uploadToken     string
	subSecret       string
	subscribe       string
	backendURL      string
	remoteConfigURL string
	httpClient      *http.Client
}

func resolveSettings(settings Settings) resolvedSettings {
	dataDir := settings.DataDir
	if dataDir == "" {
		dataDir = DefaultDataDir
	}
	dataDir = filepath.Clean(dataDir)

	resolved := resolvedSettings{
		address:         firstNonEmpty(settings.Address, DefaultAddress),
		dataDir:         dataDir,
		configDir:       firstNonEmpty(settings.ConfigDir, filepath.Join(dataDir, DefaultConfigDirName)),
		uploadSecret:    firstNonEmpty(settings.UploadSecret, filepath.Join(dataDir, DefaultUploadSecret)),
		uploadToken:     firstNonEmpty(settings.UploadToken, filepath.Join(dataDir, DefaultUploadToken)),
		subSecret:       firstNonEmpty(settings.SubSecret, filepath.Join(dataDir, DefaultSubSecret)),
		subscribe:       firstNonEmpty(settings.Subscribe, filepath.Join(dataDir, DefaultSubscribe)),
		backendURL:      firstNonEmpty(settings.BackendURL, DefaultBackendURL),
		remoteConfigURL: firstNonEmpty(settings.RemoteConfigURL, DefaultRemoteConfigURL),
		httpClient:      settings.HTTPClient,
	}
	if resolved.httpClient == nil {
		resolved.httpClient = &http.Client{Timeout: DefaultHTTPTimeoutSec * time.Second}
	}
	return resolved
}

func (settings resolvedSettings) Address() string {
	return settings.address
}

func getenv(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
