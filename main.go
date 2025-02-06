package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	astgrep "github.com/lambdasawa/sarifw/pkg/ast-grep"
	"github.com/lambdasawa/sarifw/pkg/ripgrep"
	"github.com/mattn/go-shellwords"
)

type Config struct {
	OutputDir string `env:"SARIFW_OUTPUT_DIR" envDefault:".sarifw/"`
	Editor    string `env:"SARIFW_EDITOR" envDefault:"code"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg Config) error {
	sarif, err := execute()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfg.OutputDir), 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(cfg.OutputDir, "*.sarif")
	if err != nil {
		return err
	}
	log.Printf("tmp: %s", tmp.Name())

	if err := os.WriteFile(tmp.Name(), []byte(sarif), 0644); err != nil {
		return err
	}

	editorCmd, err := shellwords.Parse(cfg.Editor)
	if err != nil {
		return err
	}

	cmd := exec.Command(editorCmd[0], append(editorCmd[1:], tmp.Name())...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func execute() (string, error) {
	osArgs := os.Args[1:]
	name := osArgs[0]

	switch name {
	case "rg":
		return ripgrep.Exec(osArgs[1:])
	case "sg", "ast-grep":
		return astgrep.Exec(osArgs[1:])
	default:
		return "", fmt.Errorf("unsupported command: %s", name)
	}
}
