package command

import (
	"flag"
	"fmt"
	"iobeam.com/iobeam/client"
)

func NewTokensCommand() *Command {
	cmd := &Command{
		Name:  "token",
		Usage: "Get token for user or device",
		SubCommands: Mux{
			"user": newGetUserTokenCmd(),
			"proj": newGetProjectTokenCmd(),
		},
	}

	return cmd
}

type basicAuthData struct {
	username string
	password string
}

func (t *basicAuthData) IsValid() bool {
	return len(t.username) > 0 && len(t.password) > 0
}

func newGetUserTokenCmd() *Command {

	t := new(basicAuthData)

	cmd := &Command{
		Name:    "user",
		ApiPath: "/v1/tokens/user",
		Usage:   "get a new user token",
		Data:    t,
		Flags:   flag.NewFlagSet("tokens", flag.ExitOnError),
		Action:  getUserToken,
	}

	cmd.Flags.StringVar(&t.username, "username", "", "The username (REQUIRED)")
	cmd.Flags.StringVar(&t.password, "password", "", "The password (REQUIRED)")

	return cmd
}

func getUserToken(c *Command, ctx *Context) error {

	t := c.Data.(*basicAuthData)

	_, err := ctx.Client.
		Get(c.ApiPath).
		BasicAuth(t.username, t.password).
		Expect(200).
		ResponseBody(new(client.AuthToken)).
		ResponseBodyHandler(func(token interface{}) error {

		authToken := token.(*client.AuthToken)
		err := authToken.Save()

		if err != nil {
			fmt.Printf("Could not save token: %s\n", err)
		}

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
		Name:    "proj",
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
