package command

import (
	"bufio"
	"fmt"
	"os"

	"github.com/iobeam/iobeam/client"
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
		Usage:   "Login / switch active user.",
		Data:    t,
		Action:  getUserToken,
	}
	flags := cmd.NewFlagSet("iobeam user login")
	flags.StringVar(&t.username, "username", "", "Login username or email")

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

	var userId uint64
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
		userId = authToken.UserId

		fmt.Println("Token acquired:")
		fmt.Printf("%s\n", authToken.Token)
		return err
	}).Execute()

	// No errors, so let's set the active user id/email
	if err == nil {
		rsp := new(userData)
		_, err = ctx.Client.
			Get("/v1/users/me").
			UserToken(ctx.Profile).
			Expect(200).
			ResponseBody(rsp).
			ResponseBodyHandler(func(body interface{}) error {

			err := ctx.Profile.UpdateActiveUser(userId, rsp.Email)
			if err != nil {
				fmt.Printf("Could not update active user, you may have to login again: %s\n", err)
			}
			return err
		}).Execute()
	}

	return err
}

func printProjectToken(projToken *client.AuthToken) {
	fmt.Printf("\n%v\n\n", projToken.Token)
	fmt.Printf("Expires: %v\n", projToken.Expires)
	fmt.Printf("Permissions:\n")
	fmt.Printf("  READ  = %v\n", projToken.Read)
	fmt.Printf("  WRITE = %v\n", projToken.Write)
	fmt.Printf("  ADMIN = %v\n", projToken.Admin)
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
		Action:  getProjectToken,
	}
	flags := cmd.NewFlagSet("iobeam project token")
	flags.Uint64Var(&p.projectId, "id", ctx.Profile.ActiveProject, "Project ID (if omitted, defaults to active project)")
	flags.BoolVar(&p.read, "read", false, "Read permission")
	flags.BoolVar(&p.write, "write", false, "Write permission")
	flags.BoolVar(&p.admin, "admin", false, "Admin permissions")

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

		printProjectToken(projToken)

		return nil
	}).Execute()

	return err
}
