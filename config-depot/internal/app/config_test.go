package app

import (
	"path/filepath"
	"testing"
)

func TestSettingsFromEnvDefaultsToDataDir(t *testing.T) {
	t.Setenv(EnvDataDir, "")

	settings := SettingsFromEnv()

	if settings.DataDir != "data" {
		t.Fatalf("DataDir = %q", settings.DataDir)
	}
	if settings.ConfigDir != filepath.Join("data", "config") {
		t.Fatalf("ConfigDir = %q", settings.ConfigDir)
	}
	if settings.UploadSecret != filepath.Join("data", "upload_secret") {
		t.Fatalf("UploadSecret = %q", settings.UploadSecret)
	}
	if settings.UploadToken != filepath.Join("data", "upload_token") {
		t.Fatalf("UploadToken = %q", settings.UploadToken)
	}
	if settings.SubSecret != filepath.Join("data", "sub_secret") {
		t.Fatalf("SubSecret = %q", settings.SubSecret)
	}
	if settings.Subscribe != filepath.Join("data", "subscribe") {
		t.Fatalf("Subscribe = %q", settings.Subscribe)
	}
}
