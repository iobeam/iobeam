package command

import (
	"fmt"
	"flag"
	"strconv"
)


type CreateUserData struct {
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

func (u *CreateUserData) IsValid() bool {
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

	user := CreateUserData{}
	
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

	u := c.Data.(*CreateUserData)
	
	rsp, err := ctx.Client.
		Post(c.ApiPath).
		Body(&u).
		Expect(201).
		Execute();

	if err != nil {
		return err
	}

	userData := new(struct {
		UserId     uint64 `json:"user_id"`
	})

	err = rsp.Read(userData)

	if err != nil {
		return err
	}
 
	fmt.Printf("The new user ID for %s is %d\n",
		u.Email.String(),
		userData.UserId)
	
	return nil
}

type GetUserData struct {
	UserId requiredUint64 `json:"user_id"`
}

func (u *GetUserData) IsValid() bool {
	return u.UserId.IsValid()
}

func newGetUserCmd() (*Command) {

	user := GetUserData{}
	
	cmd := &Command {
		Name: "get",
		ApiPath: "/v1/users",
		Usage: "get user",
		Data: &user,
		Flags: flag.NewFlagSet("get", flag.ExitOnError),		
		Action: getUser,
	}

	cmd.Flags.Var(&user.UserId, "userId", "The ID of the user to query (REQUIRED)")
	
	return cmd
}

func getUser(c *Command, ctx *Context) error {

	u := c.Data.(*GetUserData)
	
	rsp, err := ctx.Client.
		Get(c.ApiPath + "/" + strconv.FormatUint(*u.UserId.v, 10)).
		Expect(200).
		Execute();
	
	if err != nil {
		return err
	}

	userData := new(struct {
		UserId     uint64 `json:"user_id"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
	})
	
	rsp.Read(userData)

	// TODO: print user data
	
	return nil
}
