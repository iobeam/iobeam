package command

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/iobeam/iobeam/config"
)

// NewConfigCommand returns the base 'config' command.
func NewConfigCommand() *Command {
	cmd := &Command{
		Name:  "profile",
		Usage: "Manage your CLI profiles.",
		SubCommands: Mux{
			"create": newCreateProfileCmd(),
			"delete": newDeleteProfileCmd(),
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
	fmt.Println()

	var user string
	var userEmail string
	if profile.ActiveUser == 0 {
		user = "[None]"
		userEmail = "[None]"
	} else {
		user = strconv.FormatUint(profile.ActiveUser, 10)
		userEmail = profile.ActiveUserEmail
		if len(userEmail) == 0 {
			rsp := new(userData)
			ctx.Client.
				Get("/v1/users/me").
				UserToken(ctx.Profile).
				Expect(200).
				ResponseBody(rsp).
				ResponseBodyHandler(func(body interface{}) error {
				return ctx.Profile.UpdateActiveUser(profile.ActiveUser, rsp.Email)
			}).Execute()
			userEmail = rsp.Email
		}
	}
	fmt.Println("Active user id:", user)
	fmt.Println("Active user   :", userEmail)
	fmt.Println()

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

type baseData struct {
	config.Profile
}

func (p *baseData) IsValid() bool {
	return len(p.Name) > 0
}

func newSwitchCmd() *Command {
	p := new(baseData)
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
	p := c.Data.(*baseData)
	err := config.SwitchProfile(p.Name)
	if err != nil {
		fmt.Println("[ERROR] Could not switch profile: ")
		return err
	}
	fmt.Printf("Active profile is now '%s'\n", p.Name)
	return nil
}

func newDeleteProfileCmd() *Command {
	p := new(baseData)
	cmd := &Command{
		Name:   "delete",
		Usage:  "Delete a profile.",
		Data:   p,
		Action: deleteProfile,
	}
	flags := cmd.NewFlagSet("iobeam profile delete")
	flags.StringVar(&p.Name, "name", "", "Name of profile to delete (cannot be active)")
	return cmd
}

func deleteProfile(c *Command, ctx *Context) error {
	p := c.Data.(*baseData)
	profile := ctx.Profile
	if p.Name == profile.Name {
		return errors.New("Cannot delete active profile")
	}

	err := config.DeleteProfile(p.Name)
	if err == nil {
		fmt.Printf("Profile '%s' successfully deleted\n", p.Name)
	} else {
		fmt.Println("[ERROR] Could not delete profile: ")
	}

	return err
}
