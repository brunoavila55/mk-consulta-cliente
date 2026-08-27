package config

import (
	"strings"
	"testing"
)

func TestLoadUsesExplicitMKCredentialNames(t *testing.T) {
	t.Setenv("CHATBOT_API_KEY", "a-chatbot-api-key-with-24-characters")
	t.Setenv("MK_BASE_URL", "http://177.72.80.20:8080")
	t.Setenv("MK_ALLOW_INSECURE_HTTP", "true")
	t.Setenv("MK_USER_ACCESS_TOKEN", "fixed-user-token")
	t.Setenv("MK_WEBSERVICE_COUNTER_PASSWORD", "webservice-counter-password")
	t.Setenv("MK_TEMPORARY_AUTH_TOKEN", "")

	settings, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.MKUserAccessToken != "fixed-user-token" {
		t.Errorf("MKUserAccessToken = %q", settings.MKUserAccessToken)
	}
	if settings.MKWebserviceCounterPassword != "webservice-counter-password" {
		t.Errorf("MKWebserviceCounterPassword = %q", settings.MKWebserviceCounterPassword)
	}
}

func TestLoadRequiresHTTPAcknowledgement(t *testing.T) {
	t.Setenv("CHATBOT_API_KEY", "a-chatbot-api-key-with-24-characters")
	t.Setenv("MK_BASE_URL", "http://177.72.80.20:8080")
	t.Setenv("MK_ALLOW_INSECURE_HTTP", "false")
	t.Setenv("MK_USER_ACCESS_TOKEN", "fixed-user-token")
	t.Setenv("MK_WEBSERVICE_COUNTER_PASSWORD", "webservice-counter-password")
	t.Setenv("MK_TEMPORARY_AUTH_TOKEN", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "sem criptografia") {
		t.Fatalf("Load() error = %v, want HTTP warning", err)
	}
}
