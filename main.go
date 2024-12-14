package main

//go:generate sqlc generate

import (
	"fmt"
	"os"

	"github.com/kymppi/kuura/cmd"
	kuura "github.com/kymppi/kuura/internal"
)

var (
	GitSHA string
	Branch string
)

func main() {
	cmd.GitSHA = GitSHA
	cmd.Branch = Branch

	config, err := kuura.ParseConfig()
	if err != nil {
		fmt.Println("Failed to load configuration:", err)
		os.Exit(1)
	}

	loggerConfig := kuura.LoggerConfig{
		DebugEnabled:  config.DEBUG,
		PrettyEnabled: config.GO_ENV != "production",
	}
	loggerManager := kuura.NewLogger(loggerConfig)
	logger := loggerManager.Get()

	rootCmd := cmd.NewRootCommand(config, logger)
	rootCmd.Execute()
}
