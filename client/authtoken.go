package client

import (
	"encoding/json"
	"os"
	"os/user"
	"strconv"
)

// AuthToken is a representation of both user and project tokens. It contains
// for all tokens the actual token string and when it expires. For user tokens
// it also contains the userID it pertains to. For project tokens it contains
// the projectID it pertains to, as well as its permissions.
type AuthToken struct {
	Token     string
	Expires   string
	UserId    uint64 `json:"user_id,omitempty"`
	ProjectId uint64 `json:"project_id,omitempty"`
	Read      bool   `json:",omitempty"`
	Write     bool   `json:",omitempty"`
	Admin     bool   `json:",omitempty"`
}

const dotDirName = ".iobeam"
const userTokenFile = "token.json"
const pathSeparator = string(os.PathSeparator)

func tokenDir() string {
	user, err := user.Current()

	if err != nil {
		return os.TempDir()
	}

	return user.HomeDir + pathSeparator + dotDirName
}

func userTokenPath() string {
	return tokenDir() + pathSeparator + userTokenFile
}

func projTokenPath(id uint64) string {
	return tokenDir() + pathSeparator + "proj_" + strconv.FormatUint(id, 10) + ".json"
}

// Save writes the token to disk in the user's .iobeam directory.
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

// ReadUserToken fetches the user token that is stored on disk, if it exists.
func ReadUserToken() (*AuthToken, error) {
	return readToken(userTokenPath())
}

// ReadProjToken fetches the project token for a particular id from the disk,
// if it exists.
func ReadProjToken(id uint64) (*AuthToken, error) {
	return readToken(projTokenPath(id))
}
