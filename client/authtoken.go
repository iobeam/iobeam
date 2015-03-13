package client

import (
	"os/user"
	"encoding/json"
	"os"
)

type AuthToken struct {
	Token    string
	Expires  string
	UserId   uint64 `json:"user_id"`
}


const TOKEN_FILENAME = "token.json"

func tokenDir() string {
	user, err := user.Current()

	if err != nil {
		return os.TempDir()
	}

	return user.HomeDir + "/.cerebctl"
}

func tokenFile() string {
	return tokenDir() + "/" + TOKEN_FILENAME
}

func (t *AuthToken) Save() error {

	err := os.MkdirAll(tokenDir(), 0700)

	if err != nil {
		return err
	}

	file, err := os.OpenFile(tokenFile(), os.O_CREATE|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	defer file.Close()
	
	return json.NewEncoder(file).Encode(t)
}

func (t *AuthToken) Read() error {

	file, err := os.Open(tokenFile())

	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(t)
}

func ReadToken() (*AuthToken, error) {
	t := new(AuthToken)

	err := t.Read()

	if err != nil {
		return nil, err
	}

	return t, nil
}
