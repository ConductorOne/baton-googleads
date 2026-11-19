package connector

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	googlecredentialguard "github.com/conductorone/google-credential-guard"
)

func writeCredentials(t *testing.T, contents []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "credentials.json")
	if err := os.WriteFile(path, contents, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCredentialsOptionAcceptsServiceAccount(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{
		"type":                        "service_account",
		"project_id":                  "test-project",
		"private_key_id":              "test-key",
		"private_key":                 string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})),
		"client_email":                "test@test-project.iam.gserviceaccount.com",
		"client_id":                   "123456789012345678901",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/test",
		"universe_domain":             "googleapis.com",
	}
	for _, input := range []struct {
		name       string
		tokenURI   string
		privateKey string
	}{
		{"current", googlecredentialguard.TokenURL, fields["private_key"]},
		{"console endpoint", googlecredentialguard.AccountsTokenURL, fields["private_key"]},
		{"PKCS1 key", googlecredentialguard.AccountsTokenURL,
			string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))},
	} {
		t.Run(input.name, func(t *testing.T) {
			fields["token_uri"] = input.tokenURI
			fields["private_key"] = input.privateKey
			data, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := credentialsOption(writeCredentials(t, data)); err != nil {
				t.Fatalf("valid service account rejected: %v", err)
			}
		})
	}
}

func TestCredentialsOptionRejectsUnsafeFiles(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"external account file", []byte(`{
			"type":"external_account",
			"audience":"//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/pool/providers/provider",
			"subject_token_type":"urn:ietf:params:oauth:token-type:jwt",
			"token_url":"https://example.invalid/token",
			"credential_source":{"file":"/tmp/subject-token"}
		}`)},
		{"external account URL", []byte(`{"type":"external_account","credential_source":{"url":"https://example.invalid/subject-token"}}`)},
		{"external account executable", []byte(`{"type":"external_account","credential_source":{"executable":{"command":"id"}}}`)},
		{"oversized", []byte(strings.Repeat("x", googlecredentialguard.MaxJSONBytes+1))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := credentialsOption(writeCredentials(t, tc.data)); err == nil {
				t.Fatal("unsafe credentials accepted")
			}
		})
	}
}

func TestAllGoogleAdsCallsitesRejectExternalAccount(t *testing.T) {
	path := writeCredentials(t, []byte(`{"type":"external_account","credential_source":{"file":"/tmp/subject-token"}}`))
	ctx := context.Background()
	token := &pagination.Token{}
	checks := map[string]func() error{
		"accounts.List": func() error { _, _, _, err := newAccountBuilder("token", "", path).List(ctx, nil, token); return err },
		"users.List":    func() error { _, _, _, err := newUserBuilder("token", "", path).List(ctx, nil, token); return err },
		"roles.Grants": func() error {
			_, _, _, err := newRoleBuilder("token", "", path).Grants(ctx, &v2.Resource{Id: &v2.ResourceId{Resource: "ADMIN"}}, token)
			return err
		},
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			if err := check(); err == nil || !strings.Contains(err.Error(), "validate credentials file") {
				t.Fatalf("expected service-account rejection, got %v", err)
			}
		})
	}
}
