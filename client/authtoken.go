package client

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/iobeam/iobeam/config"
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

// see time.Parse docs for why this is the case
const tokenTimeFormat = "2006-01-02 15:04:05 -0700"

const userTokenFile = "token.json"
const pathSeparator = string(os.PathSeparator)

func userTokenPath(p *config.Profile) string {
	return p.GetDir() + pathSeparator + userTokenFile
}

func projTokenPath(p *config.Profile, id uint64) string {
	return p.GetDir() + pathSeparator + "proj_" + strconv.FormatUint(id, 10) + ".json"
}

// IsExpired reports whether AuthToken t has expired.
func (t *AuthToken) IsExpired() (bool, error) {
	exp, err := time.Parse(tokenTimeFormat, t.Expires)
	if err != nil {
		return false, err
	}

	now := time.Now()
	return now.After(exp), nil
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
