package command

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/iobeam/iobeam/client"
	"os"
)

type basicAuthData struct {
	username string
	password string
}

func (t *basicAuthData) IsValid() bool {
	return true
}

func newGetUserTokenCmd() *Command {
	t := new(basicAuthData)

	cmd := &Command{
		Name:    "login",
		ApiPath: "/v1/tokens/user",
		Usage:   "Log in as a user / switch active user. If flag is not set, you will be prompted for email/username.",
		Data:    t,
		Flags:   flag.NewFlagSet("user login", flag.ExitOnError),
		Action:  getUserToken,
	}

	cmd.Flags.StringVar(&t.username, "username", "", "Login username or email")

	return cmd
}

func getUserToken(c *Command, ctx *Context) error {
	t := c.Data.(*basicAuthData)

	if len(t.username) == 0 {
		bio := bufio.NewReader(os.Stdin)
		fmt.Printf("Username/email: ")
		line, _, err := bio.ReadLine()
		if err != nil {
			return err
		}
		t.username = string(line)
	}
	// FIXME: do not echo old password
	bio := bufio.NewReader(os.Stdin)
	fmt.Printf("Password: ")
	line, _, err := bio.ReadLine()
	if err != nil {
		return err
	}
	t.password = string(line)

	_, err = ctx.Client.
		Get(c.ApiPath).
		BasicAuth(t.username, t.password).
		Expect(200).
		ResponseBody(new(client.AuthToken)).
		ResponseBodyHandler(func(token interface{}) error {

		authToken := token.(*client.AuthToken)
		err := authToken.Save(ctx.Profile)
		if err != nil {
			fmt.Printf("Could not save token: %s\n", err)
		}

		err = ctx.Profile.UpdateActiveUser(authToken.UserId)
		if err != nil {
			fmt.Printf("Could not update active user: %s\n", err)
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

func newGetProjectTokenCmd(ctx *Context) *Command {

	p := new(projectPermissions)

	cmd := &Command{
		Name:    "token",
		ApiPath: "/v1/tokens/project",
		Usage:   "Get a project token.",
		Data:    p,
		Flags:   flag.NewFlagSet("tokens", flag.ExitOnError),
		Action:  getProjectToken,
	}

	cmd.Flags.Uint64Var(&p.projectId, "id", ctx.Profile.ActiveProject, "Project ID (if omitted, defaults to active project)")
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
		ParamBool("include_user", false). // TODO: make toggleable?
		UserToken(ctx.Profile).
		Expect(200).
		ResponseBody(new(client.AuthToken)).
		ResponseBodyHandler(func(token interface{}) error {

		projToken := token.(*client.AuthToken)
		err := projToken.Save(ctx.Profile)
		if err != nil {
			fmt.Printf("Could not save token: %s\n", err)
		}

		err = ctx.Profile.UpdateActiveProject(projToken.ProjectId)
		if err != nil {
			fmt.Printf("Could not update active project: %s\n", err)
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
