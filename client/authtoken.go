package client

import (
	"encoding/json"
	"github.com/iobeam/iobeam/config"
	"os"
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

const userTokenFile = "token.json"
const pathSeparator = string(os.PathSeparator)

func userTokenPath(p *config.Profile) string {
	return p.GetDir() + pathSeparator + userTokenFile
}

func projTokenPath(p *config.Profile, id uint64) string {
	return p.GetDir() + pathSeparator + "proj_" + strconv.FormatUint(id, 10) + ".json"
}

// Save writes the token to disk in the user's .iobeam directory.
func (t *AuthToken) Save(p *config.Profile) error {
	var tokenPath string
	if t.ProjectId == 0 {
		tokenPath = userTokenPath(p)
	} else {
		tokenPath = projTokenPath(p, t.ProjectId)
	}

	dir := p.GetDir()
	err := os.MkdirAll(dir, 0700)

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
func ReadUserToken(p *config.Profile) (*AuthToken, error) {
	return readToken(userTokenPath(p))
}

// ReadProjToken fetches the project token for a particular id from the disk,
// if it exists.
func ReadProjToken(p *config.Profile, id uint64) (*AuthToken, error) {
	return readToken(projTokenPath(p, id))
}
