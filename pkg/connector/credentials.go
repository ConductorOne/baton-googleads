package connector

import (
	"fmt"
	"io"
	"os"

	googlecredentialguard "github.com/conductorone/google-credential-guard"
	"google.golang.org/api/option"
)

// credentialsOption reads the operator-configured local file with a strict size
// bound and only permits validated service-account credentials.
func credentialsOption(path string) (option.ClientOption, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open credentials file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, googlecredentialguard.MaxJSONBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read credentials file: %w", err)
	}
	key, err := googlecredentialguard.ParseServiceAccountJSON(data)
	if err != nil {
		return nil, fmt.Errorf("validate credentials file: %w", err)
	}
	return key.ClientOption()
}
