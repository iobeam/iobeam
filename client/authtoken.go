package client

import (
	"encoding/json"
	"os"
	"os/user"
	"strconv"
)

type AuthToken struct {
	Token     string
	Expires   string
	UserId    uint64 `json:"user_id,omitempty"`
	ProjectId uint64 `json:"project_id,omitempty"`
	Read      bool   `json:",omitempty"`
	Write     bool   `json:",omitempty"`
	Admin     bool   `json:",omitempty"`
}

const userTokenFile = "token.json"
const pathSeparator = string(os.PathSeparator)

func tokenDir() string {
	user, err := user.Current()

	if err != nil {
		return os.TempDir()
	}

	return user.HomeDir + pathSeparator + ".beam"
}

func userTokenPath() string {
	return tokenDir() + pathSeparator + userTokenFile
}

func projTokenPath(id uint64) string {
	return tokenDir() + pathSeparator + "proj_" + strconv.FormatUint(id, 10) + ".json"
}

func (t *AuthToken) Save() error {
	var tokenPath string
	if t.ProjectId == 0 {
		tokenPath = userTokenPath()
	} else {
		tokenPath = projTokenPath(t.ProjectId)
	}

	err := os.MkdirAll(tokenDir(), 0700)

	if err != nil {
		return err
	}

	_ = os.Remove(tokenPath)
	// Error only if it does not exist, ignore.

	file, err := os.OpenFile(tokenPath, os.O_CREATE|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	defer file.Close()

	return json.NewEncoder(file).Encode(t)
}

func (t *AuthToken) read(tokenPath string) error {

	file, err := os.Open(tokenPath)

	if err != nil {
		return err
	}

	return json.NewDecoder(file).Decode(t)
}

func readToken(tokenPath string) (*AuthToken, error) {
	t := new(AuthToken)

	err := t.read(tokenPath)
	if err != nil {
		return nil, err
	}

	return t, err
}

func ReadUserToken() (*AuthToken, error) {
	return readToken(userTokenPath())
}

func ReadProjToken(id uint64) (*AuthToken, error) {
	return readToken(projTokenPath(id))
}
