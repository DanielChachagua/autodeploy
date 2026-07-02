package deploy

import (
	"fmt"
	"os"

	"autodeploy/internal/platform/logger"
	"autodeploy/internal/platform/runner"

	"github.com/spf13/cobra"
)

var (
	configPath  string
	composePath string
	envPath     string
	dryRun      bool
	onlySteps   string
	skipSteps   string
)

// NewDeployCommand crea el comando principal de despliegue para Cobra.
func NewDeployCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "autodeploy",
		Short: "Herramienta CLI para automatizar despliegues con Docker, Nginx y Certbot.",
		Long:  `AutoDeploy es una herramienta diseñada bajo arquitectura hexagonal para automatizar la instalación de dependencias, configuración de reverse proxy en Nginx e instalación de SSL con Certbot.`,
	}

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newDestroyCmd())

	return rootCmd
}

func newRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Ejecuta el proceso completo de despliegue automático",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := logger.NewLogger()
			r := runner.NewRunner(dryRun, l.Stdout())
			repo := NewRepository()
			svc := NewService(repo, r, l)

			opts := DeployOptions{
				Only: onlySteps,
				Skip: skipSteps,
			}

			if err := svc.Deploy(configPath, composePath, envPath, opts); err != nil {
				l.Error("%v", err)
				return err
			}
			return nil
		},
	}

	runCmd.Flags().StringVarP(&configPath, "config", "c", "autodeploy.yaml", "Ruta al archivo autodeploy.yaml")
	runCmd.Flags().StringVarP(&composePath, "compose", "d", "docker-compose.yml", "Ruta al archivo docker-compose.yml")
	runCmd.Flags().StringVarP(&envPath, "env", "e", ".env", "Ruta al archivo de variables de entorno .env")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simula la ejecución de comandos sin realizarlos en el sistema")
	runCmd.Flags().StringVarP(&onlySteps, "only", "o", "", "Ejecuta solo ciertos pasos, separados por comas (valores: dependencies, nginx, ssl)")
	runCmd.Flags().StringVarP(&skipSteps, "skip", "s", "", "Omite ciertos pasos, separados por comas (valores: dependencies, nginx, ssl)")

	return runCmd
}

func newDestroyCmd() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "Revierte los cambios aplicados (configuración de Nginx y certificados SSL)",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := logger.NewLogger()
			r := runner.NewRunner(dryRun, l.Stdout())
			repo := NewRepository()
			svc := NewService(repo, r, l)

			opts := DeployOptions{
				Only: onlySteps,
				Skip: skipSteps,
			}

			if err := svc.Destroy(configPath, opts); err != nil {
				l.Error("%v", err)
				return err
			}
			return nil
		},
	}

	destroyCmd.Flags().StringVarP(&configPath, "config", "c", "autodeploy.yaml", "Ruta al archivo autodeploy.yaml")
	destroyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simula la ejecución de la reversión de comandos sin realizarlos en el sistema")
	destroyCmd.Flags().StringVarP(&onlySteps, "only", "o", "", "Revierte solo ciertos pasos, separados por comas (valores: nginx, ssl)")
	destroyCmd.Flags().StringVarP(&skipSteps, "skip", "s", "", "Omite la reversión de ciertos pasos, separados por comas (valores: nginx, ssl)")

	return destroyCmd
}

func newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Inicializa archivos de plantilla (autodeploy.yaml, docker-compose.yml, .env) en el directorio actual",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := logger.NewLogger()

			// 1. Crear autodeploy.yaml de ejemplo
			yamlTemplate := `domains:
  - ejemplo.com
  - www.ejemplo.com
email: tu-email@ejemplo.com
routes:
  - path: /api/
    service: api
    client_max_body_size: 50M
    cors: true
  - path: /dozzle/
    service: dozzle
    auth_basic: "Acceso Restringido - NOA GESTION"
    auth_basic_user_file: /etc/nginx/.htpasswd_dozzle
  - path: /
    service: web
    websocket: true
`
			if err := writeTemplate("autodeploy.yaml", yamlTemplate, l); err != nil {
				return err
			}

			// 2. Crear docker-compose.yml de ejemplo
			composeTemplate := `services:
  web:
    image: nginx:alpine
    ports:
      - "5000:80"
  api:
    image: node:alpine
    ports:
      - "${API_PORT}:3000"
  dozzle:
    image: amir20/dozzle:latest
    ports:
      - "8888:8080"
`
			if err := writeTemplate("docker-compose.yml", composeTemplate, l); err != nil {
				return err
			}

			// 3. Crear .env de ejemplo
			envTemplate := `API_PORT=3000
`
			if err := writeTemplate(".env", envTemplate, l); err != nil {
				return err
			}

			l.Success("Plantillas creadas exitosamente en el directorio de trabajo.")
			return nil
		},
	}

	return initCmd
}

func writeTemplate(filename, content string, l logger.Logger) error {
	if _, err := os.Stat(filename); err == nil {
		l.Warn("El archivo %s ya existe. Saltando...", filename)
		return nil
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo %s: %w", filename, err)
	}

	l.Info("Archivo de plantilla %s generado.", filename)
	return nil
}
