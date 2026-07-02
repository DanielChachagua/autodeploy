package deploy

// Route define el mapeo de una ruta Nginx hacia un puerto local o servicio Docker.
type Route struct {
	Path               string `yaml:"path"`
	Service            string `yaml:"service,omitempty"`
	Port               int    `yaml:"port,omitempty"`
	ProxyPass          string `yaml:"proxy_pass,omitempty"`
	ClientMaxBodySize  string `yaml:"client_max_body_size,omitempty"`
	AuthBasic          string `yaml:"auth_basic,omitempty"`
	AuthBasicUserFile  string `yaml:"auth_basic_user_file,omitempty"`
	Websocket          bool   `yaml:"websocket,omitempty"`
	CORS               bool   `yaml:"cors,omitempty"`
}

// AutoDeployConfig mapea el archivo de configuración autodeploy.yaml.
type AutoDeployConfig struct {
	Domains []string `yaml:"domains"`
	Email   string   `yaml:"email"`
	Routes  []Route  `yaml:"routes"`
}

// DockerComposeService mapea la estructura básica de un servicio en docker-compose.yml.
type DockerComposeService struct {
	Ports []any `yaml:"ports,omitempty"`
}

// DockerCompose mapea el archivo completo docker-compose.yml.
type DockerCompose struct {
	Services map[string]DockerComposeService `yaml:"services"`
}
