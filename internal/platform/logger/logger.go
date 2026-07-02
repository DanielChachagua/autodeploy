package logger

import (
	"fmt"
	"io"
	"os"
)

// Colores ANSI básicos
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// Logger define la interfaz para registrar eventos en la consola.
type Logger interface {
	Info(format string, a ...any)
	Success(format string, a ...any)
	Warn(format string, a ...any)
	Error(format string, a ...any)
	Step(format string, a ...any)
	Stdout() io.Writer
}

type consoleLogger struct {
	stdout io.Writer
	stderr io.Writer
}

// NewLogger crea un logger para la CLI.
func NewLogger() Logger {
	return &consoleLogger{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (l *consoleLogger) Stdout() io.Writer {
	return l.stdout
}

func (l *consoleLogger) Info(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(l.stdout, "%s%s%s\n", colorBlue, msg, colorReset)
}

func (l *consoleLogger) Success(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(l.stdout, "%s✨ %s%s\n", colorGreen, msg, colorReset)
}

func (l *consoleLogger) Warn(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(l.stderr, "%s⚠️  %s%s\n", colorYellow, msg, colorReset)
}

func (l *consoleLogger) Error(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(l.stderr, "%s❌ Error: %s%s\n", colorRed, msg, colorReset)
}

func (l *consoleLogger) Step(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(l.stdout, "%s🚀 %s%s\n", colorCyan+colorBold, msg, colorReset)
}
