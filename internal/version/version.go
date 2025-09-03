package version

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var (
	appVersion = "dev"
	commitHash = "unknown"
	verFlag    bool
)

type AppVersion struct {
	Release    string `json:"release"`
	CommitHash string `json:"commit_hash"`
}

func (s AppVersion) String() string {
	return fmt.Sprintf("version info:\n\n release: %s\n commit hash: %s \n go version: %s \n)",
		s.Release,
		s.CommitHash,
		runtime.Version(),
	)
}

func (s AppVersion) LogString() string {
	return fmt.Sprintf("%s-%s", s.Release, s.CommitHash)
}

func InitFlags() {
	fset := flag.NewFlagSet("version", flag.ExitOnError)
	fset.BoolVar(&verFlag, "version", false, "Print current version of the service")
	fset.BoolVar(&verFlag, "v", false, "Print current version of the service")
	if err := fset.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	Print()
}

func Print() {
	if verFlag {
		fmt.Println(Get().String())
		os.Exit(0)
	}
}

func Get() AppVersion {
	return AppVersion{
		Release:    appVersion,
		CommitHash: commitHash,
	}
}
