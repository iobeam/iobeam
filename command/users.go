package command

import (
	"fmt"
	"flag"
	"strconv"
	"bufio"
	"os"
	"beam.io/beam/client"
)

type userData struct {
	Email         requiredString `json:"email"`
	Password      requiredString `json:"password"`
	Username      string `json:"username,omitempty"`
	Url           string `json:"url,omitempty"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	CompanyName   string `json:"company_name,omitempty"`
	// Private fields, not marshalled into JSON
	emailIsSet    bool
	passwordIsSet bool
	isUpdate      bool
}

func (u *userData) IsValid() bool {
	if u.isUpdate {
		return u.Email.IsValid() ||
			u.Password.IsValid() ||
			len(u.Url) > 0 ||
			len(u.FirstName) > 0 ||
			len(u.LastName) > 0 ||
			len(u.CompanyName) > 0
	}
	return u.Email.IsValid() && u.Password.IsValid()
}

func NewUsersCommand() *Command {
	cmd := &Command {
		Name: "user",
		Usage: "Create, get, or delete users",
		SubCommands: Mux {
			"get": newGetUserCmd(),
			"create": newCreateUserCmd(),
			"update": newUpdateUserCmd(),
			"search": newSearchUsersCmd(),
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

	if (update) {
		apiPath += "/me"
		flags.StringVar(&user.Username, "username", "",
			"Username associated with user")
	}
	flags.Var(&user.Password, "password", "The user's password" +
		requiredArg(!update))
	flags.Var(&user.Email, "email", "The user's email address" +
		requiredArg(!update))
	flags.StringVar(&user.FirstName, "firstname", "", "The user's first name")
	flags.StringVar(&user.LastName, "lastname", "", "The user's last name")
	flags.StringVar(&user.CompanyName, "company", "", "The user's company name")
	flags.StringVar(&user.Url, "url", "", "The user's webpage")
	
	cmd := &Command {
		Name: name,
		ApiPath: apiPath,
		Usage: name + " user",
		Data: &user,
		Flags: flags,	
		Action: action,
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

	if u.Password.IsValid() {
		bio := bufio.NewReader(os.Stdin)
		// FIXME: do not echo old password
		fmt.Printf("Enter old password:")
		line, _, err := bio.ReadLine()

		if err != nil {
			return err
		}
		req.Param("old_password", string(line))
	}
	_, err := req.Execute();

	if err == nil {
		fmt.Println("User successfully updated")
	}
	
	return err
}

func createUser(c *Command, ctx *Context) error {

	u := c.Data.(*userData)

	userData := new(struct {
		UserId     uint64 `json:"user_id"`
	})

	_, err := ctx.Client.
		Post(c.ApiPath).
		Body(&u).
		Expect(201).
		ResponseBody(userData).
		ResponseBodyHandler(func(interface{}) error {
		
		fmt.Printf("The new user ID for %s is %d\n",
			u.Email.String(),
			userData.UserId)
		
		return nil
	}).Execute();
		
	return err
}

type getUserData struct {
	UserId   uint64 `json:"user_id,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

func (u *getUserData) IsValid() bool {
	return true
}

func newGetUserCmd() *Command {

	user := getUserData{}
	
	cmd := &Command {
		Name: "get",
		ApiPath: "/v1/users",
		Usage: "get user information (if no user info given, returns authenticated user)",
		Data: &user,
		Flags: flag.NewFlagSet("get", flag.ExitOnError),		
		Action: getUser,
	}

	cmd.Flags.Uint64Var(&user.UserId, "userId", 0, "The ID of the user to query")
	cmd.Flags.StringVar(&user.Email, "email", "", "The email of the user to query")
	cmd.Flags.StringVar(&user.Username, "username", "", "The username of the user to query")
	
	return cmd
}

func getUser(c *Command, ctx *Context) error {

	u := c.Data.(*getUserData)

	req := ctx.Client.Get(c.ApiPath)
	
	if u.UserId != 0 {
		req = ctx.Client.Get(c.ApiPath + "/" + strconv.FormatUint(u.UserId, 10))
	} else if len(u.Email) > 0 {
		req.Param("name", u.Email)
	} else if len(u.Username) > 0 {
		req.Param("name", u.Username)
	} else {
		req = ctx.Client.Get(c.ApiPath + "/me")
	}

	user := new(struct {
		UserId     uint64 `json:"id"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
	})

	_, err := req.
		Expect(200).
		ResponseBody(user).
		ResponseBodyHandler(func(interface{}) error {

		fmt.Printf("Username: %v\n" +
			"User ID: %v\n" +
			"Email: %v\n" +
			"First name: %v\n" +
			"Last name: %v\n",
			user.Username,
			user.UserId,
			user.Email,
			user.FirstName,
			user.LastName);
		return nil
	}).Execute();

	return err
}


type searchUsersData struct {
	Username string `json:"username,omitempty"`
}

func (u *searchUsersData) IsValid() bool {
	return len(u.Username) > 0
}

func newSearchUsersCmd() *Command {

	users := searchUsersData{}
	
	cmd := &Command {
		Name: "search",
		ApiPath: "/v1/users",
		Usage: "search for users",
		Data: &users,
		Flags: flag.NewFlagSet("get", flag.ExitOnError),		
		Action: searchUsers,
	}
	cmd.Flags.StringVar(&users.Username, "name", "", "The search string")
	
	return cmd
}


func searchUsers(c *Command, ctx *Context) error {

	u := c.Data.(*searchUsersData)

	req := ctx.Client.Get(c.ApiPath)
	
	req.Param("search", u.Username)

	user := new(struct {
		Users []struct {
			UserId     uint64 `json:"id"`
			Username   string `json:"username"`
			Email      string `json:"email"`
		}
	})

	_, err := req.
		Expect(200).
		ResponseBody(user).
		ResponseBodyHandler(func(interface{}) error {

		for _, u := range(user.Users) {
			fmt.Printf("\nUsername: %v\n" +
				"User ID: %v\n" +
				"Email: %v\n",
				u.Username,
				u.UserId,
				u.Email)
			
		}
		return nil
	}).Execute();

	return err
}
