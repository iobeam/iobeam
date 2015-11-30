package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
)

type iobeamConfig struct {
	Name string `json:"profile"`
}

const (
	CLI_VERSION      = "0.3.0"
	pathSeparator    = string(os.PathSeparator)
	dotDirName       = ".iobeam"
	defaultConfig    = "profile"
	profileFileName  = "profile.config"
	DefaultApiServer = "https://api.iobeam.com"
)

func InitConfig() (*iobeamConfig, error) {
	c := &iobeamConfig{
		Name: "default",
	}
	err := c.save()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getDotDir() string {
	user, err := user.Current()

	if err != nil {
		// We cannot gracefully use the temp directory with profiles.
		panic(err)
	}

	return user.HomeDir + pathSeparator + dotDirName
}

func defaultConfigPath() string {
	return getDotDir() + pathSeparator + defaultConfig
}

func makeAllOnPath(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}
	return nil
}

func saveJson(path string, obj interface{}) error {
	_ = os.Remove(path) // error only if it does not exist, ignore.

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer file.Close()
	return json.NewEncoder(file).Encode(obj)
}

func readJson(path string, obj interface{}) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}

	defer file.Close()
	return json.NewDecoder(file).Decode(obj)
}

// save writes the config to disk in the user's .iobeam directory.
func (c *iobeamConfig) save() error {
	err := makeAllOnPath(getDotDir())
	if err != nil {
		return err
	}
	return saveJson(defaultConfigPath(), c)
}

func (c *iobeamConfig) read(path string) error {
	return readJson(path, c)
}

func readConfig(path string) (*iobeamConfig, error) {
	c := new(iobeamConfig)

	err := c.read(path)
	if err != nil {
		return nil, err
	}

	return c, err
}

func ReadDefaultConfig() (*iobeamConfig, error) {
	return readConfig(defaultConfigPath())
}

type Profile struct {
	Name            string `json:"-"`
	Server          string `json:"server"`
	ActiveProject   uint64 `json:"active_project"`
	ActiveUser      uint64 `json:"active_user"`
	ActiveUserEmail string `json:"activer_user_email"`
	// TODO: Don't export active fields.
}

func InitProfile(name string) (*Profile, error) {
	return InitProfileWithServer(name, DefaultApiServer)
}

func InitProfileWithServer(name, server string) (*Profile, error) {
	p := &Profile{
		Name:          name,
		Server:        server,
		ActiveProject: 0,
		ActiveUser:    0,
	}
	err := p.save()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func baseProfilePath(name string) string {
	return getDotDir() + pathSeparator + name
}

func profilePath(name string) string {
	// e.x. ~user/.iobeam/default/profile.config
	return baseProfilePath(name) + pathSeparator + profileFileName
}

func (p *Profile) save() error {
	err := makeAllOnPath(p.GetDir())
	if err != nil {
		return err
	}
	return saveJson(profilePath(p.Name), p)
}

func (p *Profile) read() error {
	return readJson(p.GetFile(), p)
}

func (p *Profile) GetDir() string {
	return baseProfilePath(p.Name)
}

func (p *Profile) GetFile() string {
	return profilePath(p.Name)
}

func (p *Profile) UpdateActiveUser(uid uint64, email string) error {
	p.ActiveUser = uid
	p.ActiveUserEmail = email
	return p.save()
}

func (p *Profile) UpdateActiveProject(pid uint64) error {
	p.ActiveProject = pid
	return p.save()
}

func ReadProfile(name string) (*Profile, error) {
	p := new(Profile)
	p.Name = name

	err := p.read()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func GetProfileList() ([]string, error) {
	files, err := ioutil.ReadDir(getDotDir())
	if err != nil {
		return nil, err
	}

	list := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			list = append(list, f.Name())
		}
	}

	return list, nil
}

func SwitchProfile(name string) error {
	path := baseProfilePath(name)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("Profile '%s' does not exist", name)
	}

	if err == nil {
		c := &iobeamConfig{
			Name: name,
		}
		err = c.save()
	}
	return err
}

func DeleteProfile(name string) error {
	path := baseProfilePath(name)
	_, err := os.Stat(path)
	if err == nil {
		return os.RemoveAll(path)
	} else if os.IsNotExist(err) {
		return fmt.Errorf("Profile '%s' does not exist", name)
	} else {
		return err
	}
}
