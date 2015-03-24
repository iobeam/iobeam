package main

import (
	"beam.io/beam/client"
	"beam.io/beam/command"
	"flag"
	"fmt"
	"os"
)

type Arguments struct {
	Command  *string
	Username *string
	Password *string
}

func main() {
	cmd := &command.Command{
		Name:  os.Args[0],
		Usage: "Beam Command-Line Interface",
		Flags: flag.NewFlagSet("cerebriq", flag.ExitOnError),
		SubCommands: command.Mux{
			"user":    command.NewUsersCommand(),
			"token":   command.NewTokensCommand(),
			"project": command.NewProjectsCommand(),
			"device":  command.NewDevicesCommand(),
		},
	}

	var apiServer string

	cmd.Flags.StringVar(&apiServer, "server",
		"http://localhost:8080", "API server URI")

	ctx := &command.Context{
		Client: client.NewClient(&apiServer, "foo", "bar"),
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
