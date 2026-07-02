# AutoDeploy CLI

AutoDeploy es una herramienta de línea de comandos (CLI) desarrollada en **Go 1.25** diseñada para automatizar por completo el aprovisionamiento e instalación de dependencias de despliegue en servidores Ubuntu/Debian. 

La herramienta lee un archivo de configuración base (`autodeploy.yaml`), analiza tu archivo `docker-compose.yml` junto a su archivo `.env` para resolver los puertos de red expuestos, y configura automáticamente el servidor de Nginx como Reverse Proxy y gestiona el cifrado SSL con Certbot (Let's Encrypt).

---

## 🛠️ Características Principales

*   **Verificación e Instalación de Dependencias**: Detecta automáticamente si Docker, Nginx y Certbot están instalados en el sistema y, de no ser así, realiza su instalación siguiendo los pasos de configuración recomendados (incluyendo overrides de reinicio automático en systemd).
*   **Análisis Dinámico de Docker Compose**: Parsea archivos `docker-compose.yml` y resuelve los puertos locales asignados a los contenedores, incluso si están definidos mediante variables de entorno en un archivo `.env` (ej. `${API_PORT}`).
*   **Generación de Proxy Reverso Nginx**: Crea de forma automatizada los bloques de servidor en Nginx (`sites-available` / `sites-enabled`) con soporte para:
    *   Múltiples nombres de dominio.
    *   Mapeo de rutas (ej. `/`, `/api/`, `/dozzle/`) tanto a puertos manuales como a servicios del Compose.
    *   Límites de tamaño de subida de archivos (`client_max_body_size`).
    *   Autenticación básica HTTP (`auth_basic`).
    *   Políticas y cabeceras CORS.
    *   WebSockets (cabeceras `Upgrade` y `Connection`).
*   **Seguridad SSL Automatizada**: Solicita y configura los certificados SSL con Certbot de manera no interactiva y valida la renovación automática.
*   **Modo Simulación (`--dry-run`)**: Permite simular todo el flujo de comandos del sistema e inspeccionar la configuración de Nginx autogenerada sin aplicar cambios reales al servidor.

---

## 📂 Arquitectura del Proyecto

El proyecto está diseñado bajo una **Arquitectura Hexagonal (Puertos y Adaptadores)** adaptada para herramientas CLI en Go, facilitando la testabilidad y el desacoplamiento de la lógica del sistema operativo:

```text
├── go.mod                  # Módulo de Go y dependencias externas
├── main.go                 # Punto de entrada de la CLI
├── autodeploy.md           # Reglas y pautas del proyecto
├── internal/
│   ├── modules/
│   │   └── deploy/         # Módulo central de Despliegue
│   │       ├── port.go       # Puertos/Interfaces (IDeployRepository, IDeployService)
│   │       ├── schemas.go    # DTOs de configuración y Compose
│   │       ├── repository.go # Adaptador para cargar configuraciones e interpolar variables
│   │       ├── service.go    # Lógica de negocio (instalación de paquetes, Nginx, SSL)
│   │       ├── command.go    # Adaptador de entrada CLI (Cobra)
│   │       └── repository_test.go # Pruebas unitarias del módulo
│   └── platform/
│       ├── logger/         # Utilidad para formatear salidas y logs en la consola
│       └── runner/         # Utilidad para ejecutar comandos del sistema y simular ejecuciones
```

---

## ⚙️ Archivos de Configuración

Para ejecutar un despliegue, AutoDeploy requiere tres archivos en el directorio desde donde se ejecuta:

### 1. `autodeploy.yaml`
Define los dominios que se certificarán y las reglas de ruteo de Nginx.

```yaml
domains:
  - ejemplo.com
  - www.ejemplo.com
email: tu-email@ejemplo.com # Para las notificaciones de Let's Encrypt
routes:
  # Redirige /api/ al servicio 'api' definido en el docker-compose.yml
  - path: /api/
    service: api
    client_max_body_size: 50M
    cors: true # Inyecta la cabecera OPTIONS y políticas CORS básicas

  # Redirige /dozzle/ con Basic Auth hacia el contenedor 'dozzle'
  - path: /dozzle/
    service: dozzle
    auth_basic: "Acceso Restringido - NOA GESTION"
    auth_basic_user_file: /etc/nginx/.htpasswd_dozzle

  # Redirige / al contenedor 'web' y habilita WebSockets
  - path: /
    service: web
    websocket: true
```

### 2. `docker-compose.yml`
Define tus contenedores y puertos. AutoDeploy mapea el campo `service` de la ruta al puerto host publicado en el Compose.

```yaml
services:
  web:
    image: nginx:alpine
    ports:
      - "5000:80" # AutoDeploy redirigirá el path "/" a localhost:5000
  api:
    image: node:alpine
    ports:
      - "${API_PORT}:3000" # AutoDeploy interpolará el puerto usando .env
  dozzle:
    image: amir20/dozzle:latest
    ports:
      - "8888:8080"
```

### 3. `.env`
Contiene las variables utilizadas en tu archivo compose.

```env
API_PORT=3000
```

---

## 🚀 Instalación y Uso

### 1. Instalación en Servidores Linux (Binario Precompilado)
Para instalar la CLI de forma rápida en cualquier servidor Linux (sin necesidad de tener Go instalado), ejecuta el siguiente comando:

```bash
curl -sSL https://raw.githubusercontent.com/usuario/autodeploy/main/install.sh | bash
```
> [!NOTE]
> Asegúrate de reemplazar `usuario/autodeploy` en el comando y en el script `install.sh` con el repositorio real cuando esté publicado.

---

### 2. Instalación para Desarrollo (Compilación desde Código Fuente)
Si deseas modificar el código o compilarlo tú mismo, necesitas tener **Go 1.25+** instalado:

#### Usando el Makefile (Recomendado):
```bash
# Compilar el binario local
make build

# Instalar de forma global en /usr/local/bin (requiere sudo)
sudo make install

# Limpiar los binarios generados
make clean
```

#### Para crear binarios optimizados listos para distribuir (Releases):
```bash
make release
```
Esto creará ejecutables estáticos optimizados (`-ldflags="-w -s"`) en el directorio `dist/` para arquitecturas Linux `amd64` y `arm64`.

---

### 3. Inicialización
Genera automáticamente plantillas base de los archivos de configuración en tu directorio actual si aún no existen:
```bash
autodeploy init
```

### 4. Ejecución en Modo Simulación (Recomendado)
Valida y visualiza la configuración resultante de Nginx y los comandos del sistema operativo que se ejecutarán sin aplicarlos:
```bash
autodeploy run --dry-run
```

### 5. Ejecución Real en el Servidor
Para instalar las dependencias del sistema, escribir archivos en `/etc/nginx` y certificar con Certbot, la herramienta requiere ejecutarse con privilegios elevados (`sudo`):
```bash
sudo autodeploy run
```

### Opciones y Flags Adicionales del Comando `run`:
*   `-c, --config`: Ruta personalizada para `autodeploy.yaml` (Por defecto: `autodeploy.yaml`).
*   `-d, --compose`: Ruta personalizada para `docker-compose.yml` (Por defecto: `docker-compose.yml`).
*   `-e, --env`: Ruta personalizada para el archivo `.env` (Por defecto: `.env`).
*   `--dry-run`: Simula la ejecución sin realizar cambios reales en el sistema operativo.

---

## 🧪 Pruebas Automatizadas

Puedes verificar el correcto funcionamiento del parseador, la interpolación de variables de entorno y el mapeo de puertos ejecutando las pruebas unitarias:

```bash
go test -v ./...
```
