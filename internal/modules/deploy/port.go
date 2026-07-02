package deploy

// IDeployRepository define las operaciones para leer y resolver configuraciones.
type IDeployRepository interface {
	LoadConfig(path string) (*AutoDeployConfig, error)
	LoadCompose(path string) (*DockerCompose, error)
	ResolvePort(serviceName string, compose *DockerCompose, envMap map[string]string) (string, error)
	LoadEnvFile(path string) (map[string]string, error)
	InterpolateString(str string, envMap map[string]string) string
}

// DeployOptions define opciones para filtrar los pasos que se ejecutan.
type DeployOptions struct {
	Only string // Pasos a ejecutar (ej. "dependencies", "nginx", "ssl")
	Skip string // Pasos a ignorar (ej. "dependencies", "nginx", "ssl")
}

// IDeployService define la lógica de negocio para ejecutar las tareas de despliegue.
type IDeployService interface {
	CheckAndInstallDependencies() error
	Deploy(configPath, composePath, envPath string, opts DeployOptions) error
	Destroy(configPath string, opts DeployOptions) error
}
