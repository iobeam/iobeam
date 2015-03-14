package command

import (
	"cerebriq.com/cerebctl/client"
	"fmt"
	"flag"
)

func NewTokensCommand() (*Command) {
	cmd := &Command {
		Name: "token",
		Usage: "Get new user or device tokens",
		SubCommands: Mux {
			"get": newGetUserTokenCmd(),
		},
	}

	return cmd
}

type tokenRequestData struct {
	username requiredString
	password requiredString
}

func (t *tokenRequestData) IsValid() bool {
	return t.username.IsValid() && t.password.IsValid()
}

func newGetUserTokenCmd() (*Command) {

	t := new(tokenRequestData)
	
	cmd := &Command {
		Name: "get",
		ApiPath: "/v1/tokens/user",
		Usage: "get a new user token",
		Data: t,
		Flags: flag.NewFlagSet("tokens", flag.ExitOnError),		
		Action: getUserToken,
	}

	cmd.Flags.Var(&t.username, "username", "The username (REQUIRED)")
	cmd.Flags.Var(&t.password, "password", "The password (REQUIRED)")

	return cmd
}

func getUserToken(c *Command, ctx *Context) error {

	t := c.Data.(*tokenRequestData)

	_, err := ctx.Client.
		Get(c.ApiPath).
		BasicAuth(t.username.String(), t.password.String()).
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
	}).Execute();
	
	return err
}


