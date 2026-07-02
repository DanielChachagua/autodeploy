package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Runner define la interfaz para ejecutar comandos del sistema.
type Runner interface {
	Run(command string, args ...string) error
	RunInDir(dir, command string, args ...string) error
	RunWithOutput(command string, args ...string) (string, error)
	IsDryRun() bool
}

type systemRunner struct {
	dryRun bool
	logger io.Writer
}

// NewRunner crea un nuevo ejecutor de comandos.
func NewRunner(dryRun bool, logger io.Writer) Runner {
	if logger == nil {
		logger = os.Stdout
	}
	return &systemRunner{
		dryRun: dryRun,
		logger: logger,
	}
}

func (r *systemRunner) IsDryRun() bool {
	return r.dryRun
}

func (r *systemRunner) Run(command string, args ...string) error {
	return r.RunInDir("", command, args...)
}

func (r *systemRunner) RunInDir(dir, command string, args ...string) error {
	fullCmd := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	if r.dryRun {
		fmt.Fprintf(r.logger, "[SIMULACIÓN] Ejecutando: %s\n", fullCmd)
		return nil
	}

	cmd := exec.Command(command, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = r.logger
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (r *systemRunner) RunWithOutput(command string, args ...string) (string, error) {
	fullCmd := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	if r.dryRun {
		fmt.Fprintf(r.logger, "[SIMULACIÓN] Ejecutando y capturando: %s\n", fullCmd)
		return "simulation-output", nil
	}

	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error al ejecutar %q: %w (stderr: %s)", fullCmd, err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}
