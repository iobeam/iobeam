package command

import (
	"fmt"
	"flag"
	"strconv"
)


type createUserData struct {
	Email         requiredString `json:"email"`
	Password      requiredString `json:"password"`
	Username      string `json:"username"`
	Url           string `json:"url"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	CompanyName   string `json:"company_name"`
	emailIsSet    bool
	passwordIsSet bool
}

func (u *createUserData) IsValid() bool {
	return u.Email.IsValid() && u.Password.IsValid()
}

func NewUsersCommand() (*Command) {
	cmd := &Command {
		Name: "user",
		Usage: "Create, get, or delete users",
		SubCommands: Mux {
			"get": newGetUserCmd(),
			"create": newCreateUserCmd(),
		},
	}

	return cmd
}

func newCreateUserCmd() (*Command) {

	user := createUserData{}
	
	cmd := &Command {
		Name: "create",
		ApiPath: "/v1/users",
		Usage: "create user",
		Data: &user,
		Flags: flag.NewFlagSet("create", flag.ExitOnError),	
		Action: createUser,
	}

	cmd.Flags.Var(&user.Password, "password", "The user's password (REQUIRED)")
	cmd.Flags.Var(&user.Email, "email", "The user's email address (REQUIRED)")
	cmd.Flags.StringVar(&user.Username, "username", "", "Username associated with user")
	cmd.Flags.StringVar(&user.FirstName, "firstname", "", "The user's first name")
	cmd.Flags.StringVar(&user.LastName, "lastname", "", "The user's last name")
	cmd.Flags.StringVar(&user.CompanyName, "company", "", "The user's company name")
	cmd.Flags.StringVar(&user.Url, "url", "", "The user's webpage")
	return cmd
}

func createUser(c *Command, ctx *Context) error {

	u := c.Data.(*createUserData)

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
	UserId   uint64 `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (u *getUserData) IsValid() bool {
	return u.UserId != 0 || len(u.Email) > 0 || len(u.Username) > 0
}

func newGetUserCmd() (*Command) {

	user := getUserData{}
	
	cmd := &Command {
		Name: "get",
		ApiPath: "/v1/users",
		Usage: "get user information (requires one of userId, email, or username)",
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
		req.Param("email", u.Email)
	} else if len(u.Username) > 0 {
		req.Param("name", u.Username)
	}

	user := new(struct {
		UserId     uint64 `json:"user_id"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
	})

	_, err := req.
		Expect(200).
		ResponseBody(user).
		ResponseBodyHandler(func(interface{}) error {

		fmt.Printf("User ID: %v\n" +
			"Username: %v\n" +
			"Email: %v\n" +
			"First name: %v\n" +
			"Last name: %v\n",
			user.UserId,
			user.Username,
			user.Email,
			user.FirstName,
			user.LastName);
		return nil
	}).Execute();

	return err
}
