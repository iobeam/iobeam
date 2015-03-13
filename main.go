
package main

import (
	"fmt"
	"os"
	"flag"
	"watchvast.com/cerebctl/command"
	"watchvast.com/cerebctl/client"
)

type Arguments struct {
	Command *string
	Username *string
	Password *string
}

var (
	cmd = &command.Command {
		Name: os.Args[0],
		Usage: "IoT client",
		SubCommands: command.Mux {
			"user": command.NewUsersCommand(),
			"token": command.NewTokensCommand(),
		},
	}
)


func main() {

	ctx := &command.Context {
		Client: client.NewClient("http://localhost:8080", "foo", "bar"),
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
