package command

import (
	"bufio"
	"fmt"
	"github.com/iobeam/iobeam/client"
	"os"
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
	isUpdate   bool
	isGet      bool
	isSearch   bool
	activeUser uint64
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
		return u.activeUser != 0 || len(u.Username) > 0
	} else if u.isSearch {
		return len(u.Username) > 0
	} else { // create new user
		return len(u.Email) > 0 && len(u.Password) > 0 && len(u.Invite) > 0
	}
}

// NewUsersCommand returns the base 'user' command.
func NewUsersCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "user",
		Usage: "Commands for managing users.",
		SubCommands: Mux{
			"create":       newCreateUserCmd(),
			"get":          newGetUserCmd(ctx),
			"login":        newGetUserTokenCmd(),
			"reset-pw":     newNewPasswordCmd(),
			"update":       newUpdateUserCmd(),
			"verify-email": newVerifyEmailCmd(),
		},
	}
	cmd.NewFlagSet("iobeam user")

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
	apiPath := "/v1/users"
	if update {
		apiPath += "/me"
	}

	cmd := &Command{
		Name:    name,
		ApiPath: apiPath,
		Usage:   name + " user",
		Data:    &user,
		Action:  action,
	}

	flags := cmd.NewFlagSet("iobeam user " + name)
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
		UserToken(ctx.Profile).
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

type getData struct {
	self     bool
	username string
}

func (d *getData) IsValid() bool {
	return d.self || len(d.username) > 0
}

func newGetUserCmd(ctx *Context) *Command {

	user := getData{
		self: (ctx.Profile.ActiveUser != 0),
	}

	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/users",
		Usage:   "get user information",
		Data:    &user,
		Action:  getUser,
	}
	flags := cmd.NewFlagSet("iobeam user get")
	flags.StringVar(&user.username, "name", "", "Username or email of the user to query")

	return cmd
}

func getUser(c *Command, ctx *Context) error {
	user := c.Data.(*getData)
	req := ctx.Client.Get(c.ApiPath)
	if len(user.username) > 0 {
		req.Param("name", user.username)
	} else { // self lookup
		req = ctx.Client.Get(c.ApiPath + "/me")
	}

	rsp := new(userData)
	_, err := req.
		UserToken(ctx.Profile).
		Expect(200).
		ResponseBody(rsp).
		ResponseBodyHandler(func(body interface{}) error {
		user := body.(*userData)
		fmt.Printf("Username  : %v\n"+
			"User ID   : %v\n"+
			"Email     : %v\n"+
			"First name: %v\n"+
			"Last name : %v\n",
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
		Action:  searchUsers,
	}
	flags := cmd.NewFlagSet("iobeam user search")
	flags.StringVar(&user.Username, "name", "", "The search string")

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
		UserToken(ctx.Profile).
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
		Action:  verifyEmail,
	}
	flags := cmd.NewFlagSet("iobeam user verify-email")
	flags.StringVar(&email.VerifyKey, "key", "", "Verification key from email")

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

	email string
}

func (p *pwData) IsValid() bool {
	return len(p.email) > 0 || len(p.ResetKey) > 0
}

func newNewPasswordCmd() *Command {
	pw := new(pwData)

	cmd := &Command{
		Name:    "reset-pw",
		ApiPath: "/v1/users/password",
		Usage:   "change password",
		Data:    pw,
		Action:  resetPassword,
	}

	flags := cmd.NewFlagSet("iobeam user reset-pw")
	flags.StringVar(&pw.email, "email", "", "If starting: email to reset password for")
	flags.StringVar(&pw.ResetKey, "key", "", "If verifying: reset key from email")

	return cmd
}

func resetPassword(c *Command, ctx *Context) error {
	pw := c.Data.(*pwData)
	if len(pw.ResetKey) > 0 {
		return verifyResetPw(pw, ctx)
	} else {
		_, err := ctx.Client.
			Get(c.ApiPath).
			Expect(204).
			Param("reset", pw.email).
			Execute()

		if err != nil {
			return err
		}

		bio := bufio.NewReader(os.Stdin)
		fmt.Printf("Verify key: ")
		line, _, err := bio.ReadLine()
		if err != nil {
			return err
		}
		pw.ResetKey = string(line)

		return verifyResetPw(pw, ctx)
	}
}

func verifyResetPw(data *pwData, ctx *Context) error {
	bio := bufio.NewReader(os.Stdin)
	for true {
		fmt.Printf("New password: ")
		line, _, err := bio.ReadLine()
		if err != nil {
			return err
		}
		if len(line) > 0 {
			data.Password = string(line)
			break
		} else {
			fmt.Println("Password cannot be empty.")
		}
	}

	_, err := ctx.Client.
		Post("/v1/users/password").
		Expect(204).
		Body(data).
		Execute()

	if err == nil {
		fmt.Println("Password successfully changed.")
	}

	return err
}
