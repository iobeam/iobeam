
package main

import (
	"fmt"
	"os"
	"flag"
	"cerebriq.com/cerebctl/command"
	"cerebriq.com/cerebctl/client"
)

type Arguments struct {
	Command *string
	Username *string
	Password *string
}

func main() {
	cmd := &command.Command {
		Name: os.Args[0],
		Usage: "CereBriq Command-Line Interface",
		Flags: flag.NewFlagSet("cerebriq", flag.ExitOnError),	
		SubCommands: command.Mux {
			"user": command.NewUsersCommand(),
			"token": command.NewTokensCommand(),
			"project": command.NewProjectsCommand(),
		},
	}

	var apiServer string
	
	cmd.Flags.StringVar(&apiServer, "server",
		"http://localhost:8080", "API server URI")

	ctx := &command.Context {
		Client: client.NewClient(&apiServer, "foo", "bar"),
		Args: os.Args,
		Index: 0,
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
