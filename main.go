package main

import (
	"flag"
	"fmt"
	"github.com/iobeam/iobeam/client"
	"github.com/iobeam/iobeam/command"
	"github.com/iobeam/iobeam/config"
	"os"
)

func getActiveProfile() *config.Profile {
	conf, err := config.ReadDefaultConfig()

	if err != nil { // config does not exist
		conf, err = config.InitConfig()
		if err != nil {
			panic("Could not initialize empty config file.")
		}
	}
	profile, err := config.ReadProfile(conf.Name)
	if profile == nil || err != nil {
		profile, err = config.InitProfile(conf.Name)
		if err != nil {
			panic("Could not initialize profile.")
		}
	}
	return profile
}

func main() {
	profile := getActiveProfile()
	server := profile.Server

	ctx := &command.Context{
		Client:  client.NewClient(&server),
		Args:    os.Args,
		Profile: profile,
		Index:   0,
	}

	cmd := &command.Command{
		Name:  os.Args[0],
		Usage: "iobeam Command-Line Interface (CLI)",
		Flags: nil,
		SubCommands: command.Mux{
			"device":  command.NewDevicesCommand(),
			"query":   command.NewExportCommand(ctx),
			"profile": command.NewConfigCommand(),
			"project": command.NewProjectsCommand(),
			"user":    command.NewUsersCommand(),
		},
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
