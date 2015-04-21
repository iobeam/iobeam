package command

import (
	"errors"
	"flag"
	"fmt"
	"github.com/iobeam/iobeam/client"
	"github.com/iobeam/iobeam/config"
	"os"
	"sort"
)

// Data is an interface for data that is posted to API, generated from command-line input.
type Data interface {
	IsValid() bool
}

type Context struct {
	Cmd     *Command
	Client  *client.Client
	Profile *config.Profile
	Index   int
	Args    []string
}

// Mux maps a subcommand name to its Command object.
type Mux map[string]*Command

// CommandAction is a function that is called when a command is invoked.
type CommandAction func(*Command, *Context) error

// Command is a representation of a CLI command including its name, usage,
// what API path it corresponds to (if any), flags, subcommands (if any), API data, and
// action to be taken when invoked.
type Command struct {
	Name        string
	Usage       string
	ApiPath     string
	Flags       *flag.FlagSet
	SubCommands Mux
	Data        Data
	Action      CommandAction
}

func (c *Command) PrintUsage() {

	if c.SubCommands != nil {
		fmt.Fprintf(os.Stderr, "Usage: %s COMMAND [FLAGS]\n\n", c.Name)
		fmt.Fprintf(os.Stderr, "%s\n", c.Usage)

		fmt.Fprint(os.Stderr, "\nAvailable Commands:\n")

		subs := make([]string, 0, len(c.SubCommands))
		for k, _ := range c.SubCommands {
			subs = append(subs, k)
		}
		sort.Strings(subs)

		for _, s := range subs {
			temp := c.SubCommands[s]
			fmt.Fprintf(os.Stderr, "  %-20s :: %s\n",
				temp.Name, temp.Usage)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Usage: %s [FLAGS]\n\n", c.Name)
		fmt.Fprintf(os.Stderr, "%s\n", c.Usage)
	}

	if c.Flags != nil {
		fmt.Fprint(os.Stderr, "\nAvailable Flags:\n")
		c.Flags.PrintDefaults()
	}
}

func (c *Command) isValid() bool {
	if c.Data == nil {
		return true
	}

	return c.Data.IsValid()
}

// Execute invokes the command it is called on.
func (c *Command) Execute(ctx *Context) error {

	ctx.Index++

	c.parseFlags(ctx)

	if c.isValid() {
		if c.Action != nil {
			return c.Action(c, ctx)
		}

		if c.SubCommands != nil && ctx.Index < len(ctx.Args) {
			sc := c.SubCommands[ctx.Args[ctx.Index]]

			if sc == nil {
				return errors.New("Invalid command '" +
					ctx.Args[ctx.Index] + "'")
			}

			return sc.Execute(ctx)
		}
	}
	c.PrintUsage()

	return nil
}

func (c *Command) parseFlags(ctx *Context) {
	if c.Flags != nil {
		c.Flags.Parse(ctx.Args[ctx.Index:])
		ctx.Index += c.Flags.NFlag()
	}
}
