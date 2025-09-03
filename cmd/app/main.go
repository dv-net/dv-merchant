package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/dv-net/dv-merchant/cmd/console"

	"github.com/urfave/cli/v2"
)

var (
	appName    = "github.com/dv-net/dv-merchant"
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
	application := &cli.App{
		Name:                 appName,
		Description:          "This is an API for DV Merchant",
		Version:              getBuildVersion(),
		Suggest:              true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			cli.HelpFlag,
			cli.VersionFlag,
			cli.BashCompletionFlag,
		},
		Commands: console.InitCommands(version, commitHash),
	}
	if err := application.Run(os.Args); err != nil {
		_, _ = fmt.Println(err.Error())
		os.Exit(1)
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
