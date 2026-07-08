package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/dv-net/dv-merchant/cmd/console"
	"github.com/dv-net/mx/logger"

	"github.com/urfave/cli/v3"
)

var (
	appName    = "dv-merchant"
	version    = "local"
	commitHash = "unknown"
)

//	@title			DV Merchant
//	@version		1.0
//	@description	This is an API for DV Merchant

//	@contact.name	DV Support
//	@contact.email	support@dv.net

//	@BasePath					/api
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

// @securityDefinitions.apikey	XApiKey
// @in							header
// @name						X-Api-Key
// @description				Store API key
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), []os.Signal{
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
	}...)
	defer cancel()

	l := logger.NewExtended()

	application := &cli.Command{
		Name:        appName,
		Description: "This is an API for DV Merchant",
		Version:     getBuildVersion(),
		Suggest:     true,
		Flags: []cli.Flag{
			cli.HelpFlag,
			cli.VersionFlag,
		},
		Commands:       console.InitCommands(version, commitHash),
		DefaultCommand: "start",
	}

	if err := application.Run(ctx, os.Args); err != nil {
		l.Fatalf("failed to run cli runner: %s", err)
	}
}

func getBuildVersion() string {
	return fmt.Sprintf(
		"\n\nrelease: %s\ncommit hash: %s\ngo version: %s",
		version,
		commitHash,
		runtime.Version(),
	)
}
