package main

import (
	"beam.io/beam/client"
	"beam.io/beam/command"
	"flag"
	"fmt"
	"os"
)

const defaultServer = "https://iot.watchvast.com"

func main() {
	cmd := &command.Command{
		Name:  os.Args[0],
		Usage: "iobeam Command-Line Interface",
		Flags: flag.NewFlagSet("iobeam", flag.ExitOnError),
		SubCommands: command.Mux{
			"user":    command.NewUsersCommand(),
			"token":   command.NewTokensCommand(),
			"project": command.NewProjectsCommand(),
			"device":  command.NewDevicesCommand(),
			"export":  command.NewExportCommand(),
		},
	}

	var apiServer string

	cmd.Flags.StringVar(&apiServer, "server",
		defaultServer, "API server URI")

	ctx := &command.Context{
		Client: client.NewClient(&apiServer),
		Args:   os.Args,
		Index:  0,
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	if len(ctx.Args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", cmd.Name)
		return
	}

	err := cmd.Execute(ctx)

	if err != nil {
		fmt.Println(err)
	}
}
