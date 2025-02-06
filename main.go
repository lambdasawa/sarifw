package main

import (
	"fmt"
	"log"
	"os"

	astgrep "github.com/lambdasawa/sarifw/pkg/ast-grep"
	"github.com/lambdasawa/sarifw/pkg/ripgrep"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	stdout, err := execAndConvert()
	if err != nil {
		return err
	}
	fmt.Println(stdout)

	return nil
}

func execAndConvert() (string, error) {
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
