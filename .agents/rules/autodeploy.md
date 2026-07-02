---
trigger: always_on
---

# Reglas y Pautas de Desarrollo - AutoDeploy (CLI)

Este archivo contiene el contexto del proyecto, la arquitectura, los estándares de codificación y las reglas de diseño que **deben ser seguidas estrictamente por cualquier asistente de IA o desarrollador** al realizar cambios en esta base de código de la herramienta CLI.

---

## 1. Arquitectura del Proyecto

El proyecto está diseñado bajo una **Arquitectura Hexagonal (Puertos y Adaptadores)** adaptada para una herramienta de línea de comandos (CLI).

*   **Capas de Dominio y Módulos (`internal/modules/...`)**:
    *   Cada módulo debe tener una estructura limpia y orientada a CLI:
        *   `port.go`: Define las interfaces de entrada y salida del módulo (ej. `IDeployRepository`, `IDeployService`).
        *   `repository.go`: Adaptador para interactuar con sistemas externos, APIs, archivos de configuración (YAML/JSON) o mockeo en memoria.
        *   `service.go`: Contiene la lógica de negocio pura del comando (ej. cómo estructurar el despliegue, validación de estado).
        *   `command.go` o `handler.go`: Adaptador de entrada CLI (ej. usando Cobra o el paquete `flag`). Se encarga de parsear los argumentos/flags de la consola, llamar al servicio y retornar el resultado o salida adecuada.
        *   `schemas.go`: Estructuras de datos (DTOs o modelos de configuración) específicos del módulo.
*   **Capas de Plataforma/Infraestructura (`internal/platform/...`)**:
    *   Contiene lógica agnóstica al dominio:
        *   `config`: Lectura y parsing de archivos de configuración global.
        *   `logger`: Utilidad de logging estructurado.
        *   `terminal`: Utilidades de formateo para la consola (colores, tablas, barras de progreso).
        *   `runner`: Utilidad para ejecutar comandos locales o remotos (SSH).

---

## 2. Gestión de Configuración y Estado

*   **Archivos de Configuración**:
    *   La herramienta debe buscar configuraciones en rutas estándar (ej. directorio actual o `~/.config/autodeploy/`).
    *   Cualquier valor sensible (passwords, llaves SSH, tokens) debe cargarse de manera segura a través de variables de entorno o archivos con permisos restrictivos (ej. `chmod 600`).
*   **Estado**:
    *   La herramienta debe ser sin estado (stateless) siempre que sea posible. Si requiere guardar el estado de los despliegues, debe hacerlo en archivos locales temporales o base de datos local ligera (como SQLite), absteniéndose de usar sistemas de caché externos a menos que se defina lo contrario.

---

## 3. Manejo de Errores, Logs y Salida por Consola

*   **Salida Estándar (`stdout`) vs Error Estándar (`stderr`)**:
    *   Toda salida informativa o resultado exitoso del comando debe enviarse a `stdout`.
    *   Los errores de ejecución, logs de depuración y warnings deben dirigirse a `stderr`.
    *   Si se implementa un flag `--json`, la salida exitosa en `stdout` debe ser estrictamente un JSON válido para facilitar el scripting y tuberías (pipes).
*   **Códigos de Salida (Exit Codes)**:
    *   Una ejecución exitosa debe finalizar siempre con código `0` (`os.Exit(0)`).
    *   Cualquier error de validación, error de entorno o fallo en ejecución debe finalizar con un código distinto de cero (ej. `os.Exit(1)`).
*   **Manejo de Errores**:
    *   Centralizar el formateo de errores que se muestran al usuario final para que sean legibles y no expongan trazas de código internas innecesarias, a menos que se use un flag de depuración (ej. `--debug` o `--verbose`).

---

## 4. Reglas de Inyección de Dependencias y Estado

*   Los servicios y repositorios deben instanciarse mediante constructores (ej. `NewService`, `NewRepository`) inyectando sus dependencias requeridas en la inicialización (ej. logger, config, runner).
*   Evitar variables globales para el estado del programa; toda configuración compartida debe pasarse de forma explícita a través de un contexto o un contenedor de dependencias simple en el punto de entrada (`cmd/`).

---

## 5. Pautas Generales de Codificación

*   Mantener el código limpio, estructurado y modular.
*   Evitar dependencias circulares.
*   Formatear siempre el código con `go fmt` antes de proponer cambios.
*   Escribir el código en español (mensajes de error de cara al usuario, comentarios explicativos generales) manteniendo el inglés para términos de negocio y del lenguaje de programación.