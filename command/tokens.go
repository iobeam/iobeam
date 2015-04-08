package command

import (
	"flag"
	"fmt"
	"iobeam.com/iobeam/client"
)

// NewTokensCommand returns the base 'token' command.
func NewTokensCommand() *Command {
	cmd := &Command{
		Name:  "token",
		Usage: "Get token for a project.",
		SubCommands: Mux{
			"project": newGetProjectTokenCmd(),
		},
	}

	return cmd
}

type basicAuthData struct {
	username string
	email    string
	password string
}

func (t *basicAuthData) IsValid() bool {
	validUsername := len(t.email) > 0 || len(t.username) > 0
	return validUsername && len(t.password) > 0
}

func newGetUserTokenCmd() *Command {

	t := new(basicAuthData)

	cmd := &Command{
		Name:    "login",
		ApiPath: "/v1/tokens/user",
		Usage:   "Log in as a user / switch active user",
		Data:    t,
		Flags:   flag.NewFlagSet("tokens", flag.ExitOnError),
		Action:  getUserToken,
	}

	cmd.Flags.StringVar(&t.username, "username", "", "Username (REQUIRED, if no -email)")
	cmd.Flags.StringVar(&t.email, "email", "", "Email (REQUIRED, if no -username)")
	cmd.Flags.StringVar(&t.password, "password", "", "Password (REQUIRED)")

	return cmd
}

func getUserToken(c *Command, ctx *Context) error {

	t := c.Data.(*basicAuthData)
	name := t.email
	if len(name) == 0 {
		name = t.username
	}

	_, err := ctx.Client.
		Get(c.ApiPath).
		BasicAuth(name, t.password).
		Expect(200).
		ResponseBody(new(client.AuthToken)).
		ResponseBodyHandler(func(token interface{}) error {

		authToken := token.(*client.AuthToken)
		err := authToken.Save()

		if err != nil {
			fmt.Printf("Could not save token: %s\n", err)
		}
		fmt.Println("Token acquired:")
		fmt.Printf("%s\n", authToken.Token)

		return err
	}).Execute()

	return err
}

type projectPermissions struct {
	projectId uint64
	read      bool
	write     bool
	admin     bool
}

func (p *projectPermissions) IsValid() bool {
	return p.projectId != 0
}

func newGetProjectTokenCmd() *Command {

	p := new(projectPermissions)

	cmd := &Command{
		Name:    "project",
		ApiPath: "/v1/tokens/project",
		Usage:   "get a new project token",
		Data:    p,
		Flags:   flag.NewFlagSet("tokens", flag.ExitOnError),
		Action:  getProjectToken,
	}

	cmd.Flags.Uint64Var(&p.projectId, "id", 0, "The project ID (REQUIRED)")
	cmd.Flags.BoolVar(&p.read, "read", false, "Read permission")
	cmd.Flags.BoolVar(&p.write, "write", false, "Write permission")
	cmd.Flags.BoolVar(&p.admin, "admin", false, "Admin permissions")

	return cmd
}

func getProjectToken(c *Command, ctx *Context) error {

	p := c.Data.(*projectPermissions)

	_, err := ctx.Client.
		Get(c.ApiPath).
		ParamUint64("project_id", p.projectId).
		ParamBool("read", p.read).
		ParamBool("write", p.write).
		ParamBool("admin", p.admin).
		Expect(200).
		ResponseBody(new(client.AuthToken)).
		ResponseBodyHandler(func(token interface{}) error {

		projToken := token.(*client.AuthToken)
		err := projToken.Save()
		if err != nil {
			fmt.Printf("Could not save token: %s\n", err)
		}

		fmt.Printf("%4v: %v\n%4v:\n  %-6v: %v\n  %-6v: %v\n  %-6v: %v\n%v\n",
			"Expires",
			projToken.Expires,
			"Permissions",
			"READ",
			projToken.Read,
			"WRITE",
			projToken.Write,
			"ADMIN",
			projToken.Admin,
			projToken.Token)

		return nil
	}).Execute()

	return err
}
