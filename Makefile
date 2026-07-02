# Makefile para AutoDeploy CLI

BINARY_NAME=autodeploy
INSTALL_DIR=/usr/bin
DIST_DIR=dist

.PHONY: all build install uninstall clean test release

all: build

build:
	@echo "Compilando AutoDeploy localmente..."
	go build -o $(BINARY_NAME) main.go
	@echo "Compilación finalizada con éxito."

install: build
	@echo "Instalando $(BINARY_NAME) en $(INSTALL_DIR)..."
	@if [ -w $(INSTALL_DIR) ]; then \
		cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME); \
		chmod +x $(INSTALL_DIR)/$(BINARY_NAME); \
	else \
		echo "Se requieren privilegios de root/sudo para instalar en $(INSTALL_DIR)"; \
		sudo cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME); \
		sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME); \
	fi
	@echo "$(BINARY_NAME) instalado correctamente. Puedes ejecutarlo con '$(BINARY_NAME)'."

uninstall:
	@echo "Desinstalando $(BINARY_NAME) de $(INSTALL_DIR)..."
	@if [ -w $(INSTALL_DIR)/$(BINARY_NAME) ]; then \
		rm -f $(INSTALL_DIR)/$(BINARY_NAME); \
	else \
		sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME); \
	fi
	@echo "$(BINARY_NAME) desinstalado correctamente."

clean:
	@echo "Limpiando binarios locales..."
	rm -f $(BINARY_NAME)
	rm -rf $(DIST_DIR)

test:
	@echo "Ejecutando pruebas unitarias..."
	go test -v ./...

release:
	@echo "Generando binarios optimizados para distribución (Linux amd64 y arm64)..."
	mkdir -p $(DIST_DIR)
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	@echo "Binarios listos en la carpeta '$(DIST_DIR)/':"
	@ls -lh $(DIST_DIR)
