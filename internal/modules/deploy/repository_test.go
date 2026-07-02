package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInterpolateString(t *testing.T) {
	repo := NewRepository()
	envMap := map[string]string{
		"PORT": "3000",
		"HOST": "localhost",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"$PORT", "3000"},
		{"${PORT}", "3000"},
		{"http://${HOST}:$PORT/api", "http://localhost:3000/api"},
		{"http://$HOST:${PORT}", "http://localhost:3000"},
		{"$UNKNOWN_VAR", "$UNKNOWN_VAR"},
		{"${UNKNOWN_VAR}", "${UNKNOWN_VAR}"},
	}

	for _, test := range tests {
		result := repo.InterpolateString(test.input, envMap)
		if result != test.expected {
			t.Errorf("Para %q esperado %q, obtenido %q", test.input, test.expected, result)
		}
	}
}

func TestLoadEnvFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "envtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	envFile := filepath.Join(tempDir, ".env")
	content := `
# Comentario
DB_PORT=5432
API_PORT="3000"
APP_NAME='AutoDeploy CLI'
EMPTY_VAL=
`
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository()
	envMap, err := repo.LoadEnvFile(envFile)
	if err != nil {
		t.Fatalf("Error cargando .env: %v", err)
	}

	if envMap["DB_PORT"] != "5432" {
		t.Errorf("DB_PORT esperado 5432, obtenido %q", envMap["DB_PORT"])
	}
	if envMap["API_PORT"] != "3000" {
		t.Errorf("API_PORT esperado 3000, obtenido %q", envMap["API_PORT"])
	}
	if envMap["APP_NAME"] != "AutoDeploy CLI" {
		t.Errorf("APP_NAME esperado 'AutoDeploy CLI', obtenido %q", envMap["APP_NAME"])
	}
}

func TestResolvePort(t *testing.T) {
	repo := NewRepository()
	envMap := map[string]string{
		"WEB_PORT": "5000",
		"API_PORT": "3000",
	}

	compose := &DockerCompose{
		Services: map[string]DockerComposeService{
			"web": {
				Ports: []any{"5000:80"},
			},
			"api": {
				Ports: []any{"${API_PORT}:3000"},
			},
			"db": {
				Ports: []any{5432},
			},
			"multi": {
				Ports: []any{"127.0.0.1:8080:80"},
			},
			"long_syntax": {
				Ports: []any{
					map[string]any{
						"target":    80,
						"published": 9000,
					},
				},
			},
		},
	}

	tests := []struct {
		service  string
		expected string
	}{
		{"web", "5000"},
		{"api", "3000"},
		{"db", "5432"},
		{"multi", "8080"},
		{"long_syntax", "9000"},
	}

	for _, test := range tests {
		port, err := repo.ResolvePort(test.service, compose, envMap)
		if err != nil {
			t.Errorf("Error resolviendo puerto para %s: %v", test.service, err)
			continue
		}
		if port != test.expected {
			t.Errorf("Servicio %s: puerto esperado %q, obtenido %q", test.service, test.expected, port)
		}
	}
}
