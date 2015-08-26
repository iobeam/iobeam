package command

import (
	"fmt"
	"github.com/iobeam/iobeam/config"
	"strconv"
)

// NewConfigCommand returns the base 'config' command.
func NewConfigCommand() *Command {
	cmd := &Command{
		Name:  "profile",
		Usage: "Manage CLI profile",
		SubCommands: Mux{
			"create": newCreateProfileCmd(),
			"info":   newProfileInfoCmd(),
			"list":   newListCmd(),
			"switch": newSwitchCmd(),
		},
	}
	cmd.NewFlagSet("iobeam profile")

	return cmd
}

type addData struct {
	config.Profile
	switchTo bool
}

func (p *addData) IsValid() bool {
	return len(p.Name) > 0 && len(p.Server) > 0
}

func newCreateProfileCmd() *Command {
	p := new(addData)
	cmd := &Command{
		Name:   "create",
		Usage:  "Create a new local profile for the CLI.",
		Data:   p,
		Action: createProfile,
	}

	flags := cmd.NewFlagSet("iobeam profile create")
	flags.StringVar(&p.Name, "name", "", "Profile name/identifier")
	flags.StringVar(&p.Server, "server", config.DefaultApiServer, "URL of API server")
	flags.BoolVar(&p.switchTo, "active", false, "Make this the active profile after creation")

	return cmd
}

func createProfile(c *Command, ctx *Context) error {
	p := c.Data.(*addData)
	_, err := config.InitProfileWithServer(p.Name, p.Server)
	fmt.Printf("Profile '%s' successfully created.\n", p.Name)
	if err == nil {
		if p.switchTo {
			err = config.SwitchProfile(p.Name)
			fmt.Printf("Active profile is now '%s'\n", p.Name)
		}
	}
	return err
}

func newProfileInfoCmd() *Command {
	cmd := &Command{
		Name:   "info",
		Usage:  "Displays the current CLI profile info.",
		Action: showInfo,
	}
	cmd.NewFlagSet("iobeam profile info")
	return cmd
}

func showInfo(c *Command, ctx *Context) error {
	profile := ctx.Profile

	fmt.Println("Profile name  :", profile.Name)
	fmt.Println("API server    :", profile.Server)
	var user string
	if profile.ActiveUser == 0 {
		user = "[None]"
	} else {
		user = strconv.FormatUint(profile.ActiveUser, 10)
	}
	fmt.Println("Active user   :", user)
	var project string
	if profile.ActiveProject == 0 {
		project = "[None]"
	} else {
		project = strconv.FormatUint(profile.ActiveProject, 10)
	}
	fmt.Println("Active project:", project)

	return nil
}

func newListCmd() *Command {
	cmd := &Command{
		Name:   "list",
		Usage:  "Displays all available profiles.",
		Action: listProfiles,
	}
	cmd.NewFlagSet("iobeam profile list")
	return cmd
}

func listProfiles(c *Command, ctx *Context) error {
	list, err := config.GetProfileList()
	profile := ctx.Profile
	if err != nil {
		return err
	}
	for _, p := range list {
		if p == profile.Name {
			fmt.Println("*", p)
		} else {
			fmt.Println(" ", p)
		}
	}
	return nil
}

type switchData struct {
	config.Profile
}

func (p *switchData) IsValid() bool {
	return len(p.Name) > 0
}

func newSwitchCmd() *Command {
	p := new(switchData)
	cmd := &Command{
		Name:   "switch",
		Usage:  "Changes the active profile.",
		Data:   p,
		Action: switchProfile,
	}
	flags := cmd.NewFlagSet("iobeam profile switch")
	flags.StringVar(&p.Name, "name", "", "Name of profile to switch to")
	return cmd
}

func switchProfile(c *Command, ctx *Context) error {
	p := c.Data.(*switchData)
	err := config.SwitchProfile(p.Name)
	if err != nil {
		fmt.Println("[ERROR] Could not switch profile: ")
		return err
	}
	fmt.Printf("Active profile is now '%s'\n", p.Name)
	return nil
}
