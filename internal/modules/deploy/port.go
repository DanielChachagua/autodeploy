package deploy

// IDeployRepository define las operaciones para leer y resolver configuraciones.
type IDeployRepository interface {
	LoadConfig(path string) (*AutoDeployConfig, error)
	LoadCompose(path string) (*DockerCompose, error)
	ResolvePort(serviceName string, compose *DockerCompose, envMap map[string]string) (string, error)
	LoadEnvFile(path string) (map[string]string, error)
	InterpolateString(str string, envMap map[string]string) string
}

// IDeployService define la lógica de negocio para ejecutar las tareas de despliegue.
type IDeployService interface {
	CheckAndInstallDependencies() error
	Deploy(configPath, composePath, envPath string) error
}
