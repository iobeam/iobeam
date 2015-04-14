package command

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/iobeam/iobeam/client"
	"os"
	"strconv"
)

type userData struct {
	Email       string `json:"email,omitempty"`
	Password    string `json:"password,omitempty"`
	Invite      string `json:"invite,omitempty"`
	UserId      uint64 `json:"user_id,omitempty"`
	Username    string `json:"username,omitempty"`
	Url         string `json:"url,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	// Private fields, not marshalled into JSON
	isUpdate bool
	isGet    bool
	isSearch bool
}

func (u *userData) IsValid() bool {
	if u.isUpdate {
		return len(u.Email) > 0 ||
			len(u.Password) > 0 ||
			len(u.Username) > 0 ||
			len(u.Url) > 0 ||
			len(u.FirstName) > 0 ||
			len(u.LastName) > 0 ||
			len(u.CompanyName) > 0
	} else if u.isGet {
		return true
	} else if u.isSearch {
		return len(u.Username) > 0
	} else { // create new user
		return len(u.Email) > 0 && len(u.Password) > 0 && len(u.Invite) > 0
	}
}

// NewUsersCommand returns the base 'user' command.
func NewUsersCommand() *Command {
	cmd := &Command{
		Name:  "user",
		Usage: "Commands for managing users.",
		SubCommands: Mux{
			"create":       newCreateUserCmd(),
			"get":          newGetUserCmd(),
			"login":        newGetUserTokenCmd(),
			"reset-pw":     newNewPasswordCmd(),
			"update":       newUpdateUserCmd(),
			"verify-email": newVerifyEmailCmd(),
		},
	}

	return cmd
}

func requiredArg(required bool) string {
	if required {
		return " (REQUIRED)"
	}
	return ""
}

func newCreateOrUpdateUserCmd(update bool, name string, action CommandAction) *Command {

	user := userData{
		isUpdate: update,
	}

	flags := flag.NewFlagSet("user", flag.ExitOnError)
	apiPath := "/v1/users"

	if update {
		apiPath += "/me"
	}
	flags.StringVar(&user.Username, "username", "",
		"Username associated with user")
	flags.StringVar(&user.Password, "password", "", "The user's password"+
		requiredArg(!update))
	flags.StringVar(&user.Email, "email", "", "The user's email address"+
		requiredArg(!update))
	flags.StringVar(&user.FirstName, "firstname", "", "The user's first name")
	flags.StringVar(&user.LastName, "lastname", "", "The user's last name")
	flags.StringVar(&user.CompanyName, "company", "", "The user's company name")
	flags.StringVar(&user.Url, "url", "", "The user's webpage")

	if !update {
		flags.StringVar(&user.Invite, "invite", "", "Invite code needed for closed beta"+
			requiredArg(true))
	}

	cmd := &Command{
		Name:    name,
		ApiPath: apiPath,
		Usage:   name + " user",
		Data:    &user,
		Flags:   flags,
		Action:  action,
	}

	return cmd
}

func newCreateUserCmd() *Command {
	return newCreateOrUpdateUserCmd(false, "create", createUser)
}

func newUpdateUserCmd() *Command {
	return newCreateOrUpdateUserCmd(true, "update", updateUser)
}

func getCreateOrUpdateRequest(ctx *Context, path string, update bool) *client.Request {
	if update {
		return ctx.Client.Patch(path)
	}
	return ctx.Client.Post(path)
}

func updateUser(c *Command, ctx *Context) error {

	u := c.Data.(*userData)

	req := ctx.Client.
		Patch(c.ApiPath).
		Body(c.Data).
		Expect(200)

	if len(u.Password) > 0 {
		bio := bufio.NewReader(os.Stdin)
		// FIXME: do not echo old password
		fmt.Printf("Enter old password: ")
		line, _, err := bio.ReadLine()

		if err != nil {
			return err
		}
		req.Param("old_password", string(line))
	}

	rsp, err := req.Execute()

	if err == nil {
		fmt.Println("User successfully updated")
	} else if rsp.Http().StatusCode == 204 {
		fmt.Println("User not modified")
		return nil
	}

	return err
}

func createUser(c *Command, ctx *Context) error {

	_, err := ctx.Client.
		Post(c.ApiPath).
		Body(c.Data).
		Expect(201).
		ResponseBody(c.Data).
		ResponseBodyHandler(func(body interface{}) error {

		u := body.(*userData)
		fmt.Printf("New user created. ID for '%s': %d\n", u.Email, u.UserId)
		fmt.Println("Acquiring token...")

		// Log new user by getting their token.
		tokenCmd := newGetUserTokenCmd()
		t := tokenCmd.Data.(*basicAuthData)
		t.username = u.Email
		t.password = u.Password
		return getUserToken(tokenCmd, ctx)
	}).Execute()

	return err
}

func newGetUserCmd() *Command {

	user := userData{
		isGet: true,
	}

	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/users",
		Usage:   "get user information",
		Data:    &user,
		Flags:   flag.NewFlagSet("get", flag.ExitOnError),
		Action:  getUser,
	}

	cmd.Flags.Uint64Var(&user.UserId, "id", 0, "The ID of the user to query")
	cmd.Flags.StringVar(&user.Username, "name", "", "The username or email of the user to query")

	return cmd
}

func getUser(c *Command, ctx *Context) error {

	user := c.Data.(*userData)

	req := ctx.Client.Get(c.ApiPath)

	if user.UserId != 0 {
		req = ctx.Client.Get(c.ApiPath + "/" + strconv.FormatUint(user.UserId, 10))
	} else if len(user.Username) > 0 {
		req.Param("name", user.Username)
	} else {
		req = ctx.Client.Get(c.ApiPath + "/me")
	}

	_, err := req.
		Expect(200).
		ResponseBody(c.Data).
		ResponseBodyHandler(func(interface{}) error {

		fmt.Printf("Username: %v\n"+
			"User ID: %v\n"+
			"Email: %v\n"+
			"First name: %v\n"+
			"Last name: %v\n",
			user.Username,
			user.UserId,
			user.Email,
			user.FirstName,
			user.LastName)
		return nil
	}).Execute()

	return err
}

func newSearchUsersCmd() *Command {

	user := userData{
		isSearch: true,
	}

	cmd := &Command{
		Name:    "search",
		ApiPath: "/v1/users",
		Usage:   "search for users",
		Data:    &user,
		Flags:   flag.NewFlagSet("get", flag.ExitOnError),
		Action:  searchUsers,
	}
	cmd.Flags.StringVar(&user.Username, "name", "", "The search string")

	return cmd
}

func searchUsers(c *Command, ctx *Context) error {

	user := new(struct {
		Users []struct {
			UserId   uint64 `json:"user_id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}
	})

	_, err := ctx.Client.
		Get(c.ApiPath).
		Param("search", c.Data.(*userData).Username).
		Expect(200).
		ResponseBody(user).
		ResponseBodyHandler(func(interface{}) error {

		for _, u := range user.Users {
			fmt.Printf("\nUsername: %v\n"+
				"User ID: %v\n"+
				"Email: %v\n",
				u.Username,
				u.UserId,
				u.Email)

		}
		return nil
	}).Execute()

	return err
}

type emailData struct {
	VerifyKey string `json:"verification_key"`
}

func (e *emailData) IsValid() bool {
	return len(e.VerifyKey) > 0
}

func newVerifyEmailCmd() *Command {
	email := new(emailData)

	cmd := &Command{
		Name:    "verify-email",
		ApiPath: "/v1/users/email",
		Usage:   "verify email",
		Data:    email,
		Flags:   flag.NewFlagSet("verify-email", flag.ExitOnError),
		Action:  verifyEmail,
	}

	cmd.Flags.StringVar(&email.VerifyKey, "key", "", "Verification key from email")

	return cmd
}

func verifyEmail(c *Command, ctx *Context) error {
	_, err := ctx.Client.
		Post(c.ApiPath).
		Expect(204).
		Body(c.Data).
		Execute()

	if err == nil {
		fmt.Println("Email successfully verified.")
	}

	return err
}

type pwData struct {
	ResetKey string `json:"reset_key"`
	Password string `json:"password"`
}

func (p *pwData) IsValid() bool {
	return len(p.ResetKey) > 0 && len(p.Password) > 0
}

func newNewPasswordCmd() *Command {
	pw := new(pwData)

	cmd := &Command{
		Name:    "reset-pw",
		ApiPath: "/v1/users/password",
		Usage:   "change password",
		Data:    pw,
		Flags:   flag.NewFlagSet("reset-pw", flag.ExitOnError),
		Action:  resetPassword,
	}

	cmd.Flags.StringVar(&pw.ResetKey, "key", "", "Reset key from email")
	cmd.Flags.StringVar(&pw.Password, "password", "", "New password")

	return cmd
}

func resetPassword(c *Command, ctx *Context) error {
	_, err := ctx.Client.
		Post(c.ApiPath).
		Expect(204).
		Body(c.Data).
		Execute()

	if err == nil {
		fmt.Println("Password successfully changed.")
	}

	return err
}
