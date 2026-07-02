package deploy

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type deployRepository struct{}

// NewRepository crea una instancia del repositorio de despliegue.
func NewRepository() IDeployRepository {
	return &deployRepository{}
}

func (r *deployRepository) LoadConfig(path string) (*AutoDeployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo de configuración: %w", err)
	}

	var config AutoDeployConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error al deserializar configuración YAML: %w", err)
	}

	return &config, nil
}

func (r *deployRepository) LoadCompose(path string) (*DockerCompose, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo docker-compose: %w", err)
	}

	var compose DockerCompose
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("error al deserializar docker-compose YAML: %w", err)
	}

	return &compose, nil
}

func (r *deployRepository) LoadEnvFile(path string) (map[string]string, error) {
	envMap := make(map[string]string)
	
	// Primero cargar las variables de entorno actuales del sistema
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Si no existe archivo .env y es opcional, retornamos lo del sistema
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return envMap, nil
		}
		return nil, fmt.Errorf("error al abrir el archivo de entorno %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignorar comentarios y líneas vacías
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Quitar comillas si existen
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}

		envMap[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error leyendo el archivo de entorno: %w", err)
	}

	return envMap, nil
}

func (r *deployRepository) InterpolateString(str string, envMap map[string]string) string {
	// Reemplaza ${VAR} y $VAR con el valor del mapa de entorno
	reBrackets := regexp.MustCompile(`\$\{([a-zA-Z0-9_]+)\}`)
	str = reBrackets.ReplaceAllStringFunc(str, func(match string) string {
		submatch := reBrackets.FindStringSubmatch(match)
		if len(submatch) > 1 {
			if val, ok := envMap[submatch[1]]; ok {
				return val
			}
		}
		return match
	})

	reNoBrackets := regexp.MustCompile(`\$([a-zA-Z0-9_]+)`)
	str = reNoBrackets.ReplaceAllStringFunc(str, func(match string) string {
		submatch := reNoBrackets.FindStringSubmatch(match)
		if len(submatch) > 1 {
			if val, ok := envMap[submatch[1]]; ok {
				return val
			}
		}
		return match
	})

	return str
}

func (r *deployRepository) ResolvePort(serviceName string, compose *DockerCompose, envMap map[string]string) (string, error) {
	if compose == nil || compose.Services == nil {
		return "", fmt.Errorf("archivo docker-compose vacío o inválido")
	}

	service, exists := compose.Services[serviceName]
	if !exists {
		return "", fmt.Errorf("el servicio %q no se encontró en docker-compose.yml", serviceName)
	}

	if len(service.Ports) == 0 {
		return "", fmt.Errorf("el servicio %q no tiene puertos expuestos", serviceName)
	}

	// Tomamos el primer puerto definido
	firstPortVal := service.Ports[0]

	switch val := firstPortVal.(type) {
	case int:
		return strconv.Itoa(val), nil
	case string:
		interpolated := r.InterpolateString(val, envMap)
		return parsePortString(interpolated)
	case map[string]any:
		// Sintaxis larga: ports: - target: 80 \n published: 8080
		if pub, ok := val["published"]; ok {
			switch pVal := pub.(type) {
			case int:
				return strconv.Itoa(pVal), nil
			case string:
				interpolated := r.InterpolateString(pVal, envMap)
				return parsePortString(interpolated)
			}
		}
	}

	return "", fmt.Errorf("tipo de puerto desconocido en el servicio %q", serviceName)
}

func parsePortString(portStr string) (string, error) {
	// Limpiar espacios
	portStr = strings.TrimSpace(portStr)

	// Sintaxis comunes:
	// 1. "3000" -> solo el puerto
	// 2. "3000:3000" -> host:container
	// 3. "127.0.0.1:3000:3000" -> ip:host:container
	parts := strings.Split(portStr, ":")
	switch len(parts) {
	case 1:
		return parts[0], nil
	case 2:
		return parts[0], nil
	case 3:
		return parts[1], nil
	default:
		return "", fmt.Errorf("formato de puerto inválido: %q", portStr)
	}
}
