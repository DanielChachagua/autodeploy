package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"autodeploy/internal/platform/logger"
	"autodeploy/internal/platform/runner"
)

type deployService struct {
	repo   IDeployRepository
	run    runner.Runner
	logger logger.Logger
}

// NewService crea una instancia del servicio de despliegue.
func NewService(repo IDeployRepository, r runner.Runner, l logger.Logger) IDeployService {
	return &deployService{
		repo:   repo,
		run:    r,
		logger: l,
	}
}

func (s *deployService) CheckAndInstallDependencies() error {
	s.logger.Step("Verificando dependencias del sistema (Docker, Nginx, Certbot)...")

	// 1. Verificar e instalar Docker
	if err := s.checkAndInstallDocker(); err != nil {
		return fmt.Errorf("fallo al configurar Docker: %w", err)
	}

	// 2. Verificar e instalar Nginx
	if err := s.checkAndInstallNginx(); err != nil {
		return fmt.Errorf("fallo al configurar Nginx: %w", err)
	}

	// 3. Verificar e instalar Certbot
	if err := s.checkAndInstallCertbot(); err != nil {
		return fmt.Errorf("fallo al configurar Certbot: %w", err)
	}

	s.logger.Success("Todas las dependencias están instaladas y configuradas.")
	return nil
}

func (s *deployService) checkAndInstallDocker() error {
	_, err := exec.LookPath("docker")
	if err == nil {
		s.logger.Success("Docker ya está instalado.")
		return nil
	}

	s.logger.Warn("Docker no está instalado. Iniciando instalación...")

	steps := [][]string{
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "ca-certificates", "curl", "gnupg", "lsb-release", "-y"},
		{"sudo", "mkdir", "-m", "0755", "-p", "/etc/apt/keyrings"},
	}

	for _, cmdArgs := range steps {
		if err := s.run.Run(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return err
		}
	}

	// Agregar la llave gpg de docker (detectando dinámicamente si es Debian o Ubuntu)
	gpgCmd := `OS_ID=$(. /etc/os-release && echo "$ID"); curl -fsSL https://download.docker.com/linux/${OS_ID}/gpg | sudo gpg --dearmor --yes -o /etc/apt/keyrings/docker.gpg`
	if s.run.IsDryRun() {
		s.logger.Info("[SIMULACIÓN] Ejecutando: %s", gpgCmd)
	} else {
		// Ejecutar a través de bash para manejar la tubería
		if err := s.run.Run("bash", "-c", gpgCmd); err != nil {
			return fmt.Errorf("error al agregar la llave GPG de Docker: %w", err)
		}
	}

	// Agregar repositorio de apt con soporte para Debian/Ubuntu y fallback para Debian testing/sid
	sourcesCmd := `OS_ID=$(. /etc/os-release && echo "$ID"); CODENAME=$(. /etc/os-release && echo "$VERSION_CODENAME"); if [ "$OS_ID" = "debian" ] && { [ "$CODENAME" = "forky" ] || [ "$CODENAME" = "trixie" ] || [ "$CODENAME" = "sid" ] || [ -z "$CODENAME" ]; }; then CODENAME="bookworm"; fi; echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/${OS_ID} ${CODENAME} stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`
	if s.run.IsDryRun() {
		s.logger.Info("[SIMULACIÓN] Ejecutando: %s", sourcesCmd)
	} else {
		if err := s.run.Run("bash", "-c", sourcesCmd); err != nil {
			return fmt.Errorf("error al agregar el repositorio de Docker: %w", err)
		}
	}

	postSteps := [][]string{
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin", "-y"},
		{"sudo", "groupadd", "docker"},
		{"sudo", "usermod", "-aG", "docker", os.Getenv("USER")},
		{"sudo", "mkdir", "-p", "/etc/systemd/system/docker.service.d/"},
	}

	for _, cmdArgs := range postSteps {
		// Ignoramos el error en groupadd docker ya que el grupo podría ya existir
		err := s.run.Run(cmdArgs[0], cmdArgs[1:]...)
		if err != nil && cmdArgs[1] != "groupadd" {
			return err
		}
	}

	// Configurar reinicio de Docker en fallos
	overrideCmd := `echo -e "[Service]\nRestart=on-failure\nRestartSec=5s" | sudo tee /etc/systemd/system/docker.service.d/override.conf`
	if s.run.IsDryRun() {
		s.logger.Info("[SIMULACIÓN] Configurando reinicio de Docker: %s", overrideCmd)
	} else {
		if err := s.run.Run("bash", "-c", overrideCmd); err != nil {
			return fmt.Errorf("error al configurar override de Docker: %w", err)
		}
	}

	systemdSteps := [][]string{
		{"sudo", "systemctl", "daemon-reload"},
		{"sudo", "systemctl", "restart", "docker"},
	}

	for _, cmdArgs := range systemdSteps {
		if err := s.run.Run(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return err
		}
	}

	s.logger.Success("Docker instalado correctamente.")
	return nil
}

func (s *deployService) checkAndInstallNginx() error {
	_, err := exec.LookPath("nginx")
	if err == nil {
		s.logger.Success("Nginx ya está instalado.")
		return nil
	}

	s.logger.Warn("Nginx no está instalado. Iniciando instalación...")

	steps := [][]string{
		{"sudo", "apt", "update"},
		{"sudo", "apt", "install", "nginx", "-y"},
		{"sudo", "systemctl", "enable", "nginx"},
		{"sudo", "mkdir", "-p", "/etc/systemd/system/nginx.service.d/"},
	}

	for _, cmdArgs := range steps {
		if err := s.run.Run(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return err
		}
	}

	// Configurar reinicio de Nginx en fallos
	overrideCmd := `echo -e "[Service]\nRestart=on-failure\nRestartSec=5s" | sudo tee /etc/systemd/system/nginx.service.d/override.conf`
	if s.run.IsDryRun() {
		s.logger.Info("[SIMULACIÓN] Configurando reinicio de Nginx: %s", overrideCmd)
	} else {
		if err := s.run.Run("bash", "-c", overrideCmd); err != nil {
			return fmt.Errorf("error al configurar override de Nginx: %w", err)
		}
	}

	systemdSteps := [][]string{
		{"sudo", "systemctl", "daemon-reload"},
		{"sudo", "systemctl", "start", "nginx"},
	}

	for _, cmdArgs := range systemdSteps {
		if err := s.run.Run(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return err
		}
	}

	s.logger.Success("Nginx instalado y configurado correctamente.")
	return nil
}

func (s *deployService) checkAndInstallCertbot() error {
	_, err := exec.LookPath("certbot")
	if err == nil {
		s.logger.Success("Certbot ya está instalado.")
		return nil
	}

	s.logger.Warn("Certbot no está instalado. Iniciando instalación...")

	steps := [][]string{
		{"sudo", "apt", "update"},
		{"sudo", "apt", "install", "certbot", "python3-certbot-nginx", "-y"},
	}

	for _, cmdArgs := range steps {
		if err := s.run.Run(cmdArgs[0], cmdArgs[1:]...); err != nil {
			return err
		}
	}

	s.logger.Success("Certbot instalado correctamente.")
	return nil
}

func (s *deployService) Deploy(configPath, composePath, envPath string, opts DeployOptions) error {
	s.logger.Step("Iniciando proceso de autodespliegue...")

	// 1. Cargar autodeploy.yaml
	config, err := s.repo.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error al cargar autodeploy.yaml: %w", err)
	}
	s.logger.Success("Configuración de autodeploy cargada correctamente.")

	// 2. Cargar entorno .env
	envMap, err := s.repo.LoadEnvFile(envPath)
	if err != nil {
		return fmt.Errorf("error al cargar el archivo de entorno .env: %w", err)
	}
	s.logger.Success("Variables de entorno cargadas.")

	// 3. Cargar docker-compose.yml (Opcional, solo requerido si se usa el campo 'service' en alguna ruta)
	var compose *DockerCompose
	if composeFile, err := s.repo.LoadCompose(composePath); err != nil {
		s.logger.Warn("No se pudo cargar docker-compose.yml: %v. (Solo se requerirá si utilizas el campo 'service' en tus rutas).", err)
	} else {
		compose = composeFile
		s.logger.Success("Archivo docker-compose.yml cargado correctamente.")
	}

	// 4. Instalar dependencias si faltan
	if s.shouldRunStep("dependencies", opts) {
		if err := s.CheckAndInstallDependencies(); err != nil {
			return err
		}
	} else {
		s.logger.Info("Saltando verificación/instalación de dependencias según filtros.")
	}

	// 5. Configurar Nginx
	if s.shouldRunStep("nginx", opts) {
		if err := s.configureNginxRoutes(config, compose, envMap); err != nil {
			return fmt.Errorf("error al configurar las rutas de Nginx: %w", err)
		}
	} else {
		s.logger.Info("Saltando configuración de Nginx según filtros.")
	}

	// 6. Configurar SSL con Certbot
	if s.shouldRunStep("ssl", opts) {
		if err := s.configureSSL(config); err != nil {
			return fmt.Errorf("error al configurar certificados SSL: %w", err)
		}
	} else {
		s.logger.Info("Saltando configuración de Certbot (SSL) según filtros.")
	}

	s.logger.Success("¡Despliegue finalizado con éxito!")
	return nil
}

func (s *deployService) configureNginxRoutes(config *AutoDeployConfig, compose *DockerCompose, envMap map[string]string) error {
	if len(config.Domains) == 0 {
		return fmt.Errorf("debes especificar al menos un dominio en la configuración")
	}

	primaryDomain := config.Domains[0]
	s.logger.Step("Generando archivo de configuración Nginx para %s...", primaryDomain)

	// Crear el bloque de servidor de Nginx
	var sb strings.Builder
	sb.WriteString("server {\n")
	sb.WriteString(fmt.Sprintf("        server_name %s;\n\n", strings.Join(config.Domains, " ")))

	for _, route := range config.Routes {
		path := route.Path
		// Limpieza y formateo de la ruta
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		sb.WriteString(fmt.Sprintf("        location %s {\n", path))

		// Client max body size
		if route.ClientMaxBodySize != "" {
			sb.WriteString(fmt.Sprintf("                client_max_body_size %s;\n", route.ClientMaxBodySize))
		}

		// Basic Auth
		if route.AuthBasic != "" && route.AuthBasicUserFile != "" {
			sb.WriteString(fmt.Sprintf("                auth_basic %q;\n", route.AuthBasic))
			sb.WriteString(fmt.Sprintf("                auth_basic_user_file %s;\n", route.AuthBasicUserFile))
		}

		// CORS block
		if route.CORS {
			sb.WriteString("                if ($request_method = 'OPTIONS') {\n")
			sb.WriteString("                        add_header 'Access-Control-Allow-Origin' '$http_origin' always;\n")
			sb.WriteString("                        add_header 'Access-Control-Allow-Methods' 'GET,POST,PUT,DELETE,OPTIONS,PATCH' always;\n")
			sb.WriteString("                        add_header 'Access-Control-Allow-Headers' 'Content-Type,Authorization,Accept,Origin,X-Requested-With,Cache-Control,Pragma' always;\n")
			sb.WriteString("                        add_header 'Access-Control-Allow-Credentials' 'true' always;\n")
			sb.WriteString("                        add_header 'Access-Control-Max-Age' '86400' always;\n")
			sb.WriteString("                        add_header 'Content-Length' '0';\n")
			sb.WriteString("                        add_header 'Content-Type' 'text/plain';\n")
			sb.WriteString("                        return 204;\n")
			sb.WriteString("                }\n")
		}

		// Determinar target de Proxy Pass
		var proxyPassTarget string
		if route.ProxyPass != "" {
			proxyPassTarget = route.ProxyPass
		} else if route.Port != 0 {
			// Usar puerto explícito en autodeploy.yaml
			proxyPassTarget = fmt.Sprintf("http://localhost:%d", route.Port)
			if path != "/" && strings.HasSuffix(path, "/") {
				proxyPassTarget = proxyPassTarget + path
			}
			s.logger.Info("Puerto explícito %d asignado a la ruta %q.", route.Port, path)
		} else if route.Service != "" {
			if compose == nil {
				return fmt.Errorf("la ruta %q requiere el servicio %q, pero no se pudo cargar docker-compose.yml (verifica que el archivo exista)", route.Path, route.Service)
			}
			port, err := s.repo.ResolvePort(route.Service, compose, envMap)
			if err != nil {
				return fmt.Errorf("no se pudo resolver el puerto para el servicio %q: %w", route.Service, err)
			}
			s.logger.Info("Servicio %q resuelto en el puerto %s.", route.Service, port)
			
			// Si la ruta contiene un path específico que queramos adjuntar, lo unimos
			// Generalmente es http://localhost:port
			proxyPassTarget = fmt.Sprintf("http://localhost:%s", port)
			if path != "/" && strings.HasSuffix(path, "/") {
				proxyPassTarget = proxyPassTarget + path
			}
		} else {
			return fmt.Errorf("la ruta %q debe definir 'proxy_pass', 'port' o 'service'", route.Path)
		}

		sb.WriteString(fmt.Sprintf("                proxy_pass %s;\n", proxyPassTarget))
		sb.WriteString("                proxy_http_version 1.1;\n")

		// Websocket support
		if route.Websocket {
			sb.WriteString("                proxy_set_header Upgrade $http_upgrade;\n")
			sb.WriteString("                proxy_set_header Connection \"upgrade\";\n")
		} else {
			sb.WriteString("                proxy_cache_bypass $http_upgrade;\n")
		}

		sb.WriteString("                proxy_set_header Host $host;\n")
		sb.WriteString("                proxy_set_header X-Real-IP $remote_addr;\n")
		sb.WriteString("                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		sb.WriteString("                proxy_set_header X-Forwarded-Proto $scheme;\n")
		sb.WriteString("                proxy_read_timeout 60s;\n")
		sb.WriteString("        }\n\n")
	}

	sb.WriteString("}\n")

	nginxConfigContent := sb.String()

	if s.run.IsDryRun() {
		s.logger.Info("Configuración de Nginx generada para simulación:\n---\n%s---", nginxConfigContent)
	}

	// Guardar la configuración temporalmente
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("nginx-%s.conf", primaryDomain))
	err := os.WriteFile(tempFile, []byte(nginxConfigContent), 0644)
	if err != nil {
		return fmt.Errorf("error al guardar archivo de configuración temporal: %w", err)
	}
	defer os.Remove(tempFile)

	s.logger.Info("Instalando configuración de Nginx para %s...", primaryDomain)

	targetPath := fmt.Sprintf("/etc/nginx/sites-available/%s", primaryDomain)
	enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", primaryDomain)

	// Copiar el archivo usando sudo
	if err := s.run.Run("sudo", "cp", tempFile, targetPath); err != nil {
		return fmt.Errorf("error al copiar configuración a sites-available (se requieren privilegios sudo): %w", err)
	}

	// Crear enlace simbólico
	if err := s.run.Run("sudo", "ln", "-sf", targetPath, enabledPath); err != nil {
		return fmt.Errorf("error al crear enlace simbólico en sites-enabled: %w", err)
	}

	// Validar configuración de Nginx
	s.logger.Info("Validando configuración de Nginx...")
	if err := s.run.Run("sudo", "nginx", "-t"); err != nil {
		return fmt.Errorf("la validación de configuración de Nginx falló: %w", err)
	}

	// Recargar Nginx
	s.logger.Info("Recargando Nginx...")
	if err := s.run.Run("sudo", "systemctl", "reload", "nginx"); err != nil {
		return fmt.Errorf("error al recargar Nginx: %w", err)
	}

	s.logger.Success("Configuración de Nginx aplicada correctamente.")
	return nil
}

func (s *deployService) configureSSL(config *AutoDeployConfig) error {
	if len(config.Domains) == 0 {
		return nil
	}

	s.logger.Step("Configurando certificado SSL con Certbot...")

	// Construir argumentos para certbot
	args := []string{"certbot", "--nginx"}
	for _, domain := range config.Domains {
		args = append(args, "-d", domain)
	}

	// Agregar modo no interactivo
	args = append(args, "--non-interactive", "--agree-tos")
	if config.Email != "" {
		args = append(args, "-m", config.Email)
	} else {
		args = append(args, "--register-unsafely-without-email")
	}

	// Ejecutar certbot
	s.logger.Info("Ejecutando certbot para los dominios: %s", strings.Join(config.Domains, ", "))
	if err := s.run.Run("sudo", args...); err != nil {
		return fmt.Errorf("error al ejecutar certbot: %w", err)
	}

	// Validar renovación (dry-run)
	s.logger.Info("Validando la renovación automática de certificados...")
	if err := s.run.Run("sudo", "certbot", "renew", "--dry-run"); err != nil {
		return fmt.Errorf("la validación de renovación de certbot falló: %w", err)
	}

	s.logger.Success("Certificados SSL configurados y validados.")
	return nil
}

// Destroy elimina los cambios aplicados (configuración de Nginx y certificados SSL).
func (s *deployService) Destroy(configPath string, opts DeployOptions) error {
	s.logger.Step("Iniciando reversión de cambios aplicados...")

	// 1. Cargar autodeploy.yaml
	config, err := s.repo.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error al cargar autodeploy.yaml: %w", err)
	}

	if len(config.Domains) == 0 {
		return fmt.Errorf("debes especificar al menos un dominio en la configuración para revertir los cambios")
	}

	primaryDomain := config.Domains[0]

	// 2. Revertir SSL con Certbot
	if s.shouldRunStep("ssl", opts) {
		s.logger.Step("Eliminando certificados SSL con Certbot para %s...", primaryDomain)
		args := []string{"certbot", "delete", "--non-interactive", "--cert-name", primaryDomain}
		s.logger.Info("Ejecutando: sudo certbot delete --non-interactive --cert-name %s", primaryDomain)
		if err := s.run.Run("sudo", args...); err != nil {
			s.logger.Warn("Advertencia al eliminar certificados: %v", err)
		} else {
			s.logger.Success("Certificados SSL eliminados.")
		}
	} else {
		s.logger.Info("Saltando eliminación de certificados SSL según filtros.")
	}

	// 3. Revertir configuración de Nginx
	if s.shouldRunStep("nginx", opts) {
		s.logger.Step("Eliminando archivos de configuración de Nginx para %s...", primaryDomain)

		targetPath := fmt.Sprintf("/etc/nginx/sites-available/%s", primaryDomain)
		enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", primaryDomain)

		// Eliminar enlace simbólico en sites-enabled
		s.logger.Info("Eliminando enlace en %s", enabledPath)
		if err := s.run.Run("sudo", "rm", "-f", enabledPath); err != nil {
			s.logger.Warn("No se pudo eliminar el enlace en sites-enabled: %v", err)
		}

		// Eliminar archivo en sites-available
		s.logger.Info("Eliminando archivo en %s", targetPath)
		if err := s.run.Run("sudo", "rm", "-f", targetPath); err != nil {
			s.logger.Warn("No se pudo eliminar el archivo en sites-available: %v", err)
		}

		// Validar configuración de Nginx
		s.logger.Info("Validando configuración de Nginx...")
		if err := s.run.Run("sudo", "nginx", "-t"); err != nil {
			s.logger.Warn("La validación de Nginx reportó advertencias: %v", err)
		}

		// Recargar Nginx
		s.logger.Info("Recargando Nginx...")
		if err := s.run.Run("sudo", "systemctl", "reload", "nginx"); err != nil {
			s.logger.Warn("No se pudo recargar Nginx: %v", err)
		} else {
			s.logger.Success("Configuración de Nginx revertida correctamente.")
		}
	} else {
		s.logger.Info("Saltando eliminación de configuración de Nginx según filtros.")
	}

	s.logger.Success("¡Reversión de cambios finalizada con éxito!")
	return nil
}

// shouldRunStep verifica si un paso específico debe ejecutarse basándose en las opciones.
func (s *deployService) shouldRunStep(step string, opts DeployOptions) bool {
	if opts.Only != "" {
		onlySteps := strings.Split(strings.ToLower(opts.Only), ",")
		for _, o := range onlySteps {
			if strings.TrimSpace(o) == step {
				return true
			}
		}
		return false
	}
	if opts.Skip != "" {
		skipSteps := strings.Split(strings.ToLower(opts.Skip), ",")
		for _, sk := range skipSteps {
			if strings.TrimSpace(sk) == step {
				return false
			}
		}
		return true
	}
	return true
}
