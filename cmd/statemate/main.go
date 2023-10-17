package main

import (
	"github.com/draganm/statemate/cmd/statemate/info"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:                 "statemate",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			info.Command(),
		},
	}
	app.RunAndExitOnError()
}
