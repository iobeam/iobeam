package command

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/iobeam/iobeam/client"
	"github.com/iobeam/iobeam/config"
)

var flagSetNames = make(map[string]string)
var baseApiPath = make(map[string]string)

// setFlags is used to call a flag multiple times to create lists or maps of
// flag values.
type setFlags map[string]struct{}

var helpFlags = map[string]bool{
	"-h":     true,
	"-help":  true,
	"--help": true,
	"--h":    true,
}

func (i *setFlags) String() string {
	return ""
}

func (i *setFlags) Set(value string) error {
	if *i == nil {
		*i = map[string]struct{}{}
	}
	var empty struct{}
	(*i)[value] = empty
	return nil
}

// Data is an interface for data that is posted to API, generated from command-line input.
type Data interface {
	IsValid() bool
}

// Context is the current CLI context.
type Context struct {
	Cmd     *Command
	Client  *client.Client
	Profile *config.Profile
	Index   int
	Args    []string
}

// Mux maps a subcommand name to its Command object.
type Mux map[string]*Command

// Action is a function that is called when a command is invoked.
type Action func(*Command, *Context) error

// Command is a representation of a CLI command including its name, usage,
// what API path it corresponds to (if any), flags, subcommands (if any), API data, and
// action to be taken when invoked.
type Command struct {
	Name        string
	Usage       string
	ApiPath     string
	flags       *flag.FlagSet
	SubCommands Mux
	Data        Data
	Action      Action
}

// NewFlagSet creates a FlagSet to use for a command, making sure it is properly
// linked to the right print usage function.
func (c *Command) NewFlagSet(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ExitOnError)
	f.Usage = func() {
		c.printUsage()
	}
	c.flags = f
	return f
}

func (c *Command) hasFlags() bool {
	return c.flags != nil
}

func (c *Command) printUsage() {
	flagsStr := ""
	if c.hasFlags() {
		flagsStr = "[FLAGS]"
	}

	if c.SubCommands != nil {
		fmt.Fprintf(os.Stderr, "Usage: %s COMMAND %s\n\n", c.Name, flagsStr)
		fmt.Fprintf(os.Stderr, "%s\n", c.Usage)

		fmt.Fprint(os.Stderr, "\nAvailable Commands:\n")

		subs := make([]string, 0, len(c.SubCommands))
		for k := range c.SubCommands {
			subs = append(subs, k)
		}
		sort.Strings(subs)

		for _, s := range subs {
			temp := c.SubCommands[s]
			fmt.Fprintf(os.Stderr, "  %-20s :: %s\n",
				temp.Name, temp.Usage)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Usage: %s %s\n\n", c.Name, flagsStr)
		fmt.Fprintf(os.Stderr, "%s\n", c.Usage)
	}

	if c.hasFlags() {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)
		c.flags.SetOutput(writer)
		c.flags.PrintDefaults()

		if writer.Buffered() > 0 {
			fmt.Fprint(os.Stderr, "\nAvailable Flags:\n")
			writer.Flush()
			b.WriteTo(os.Stderr)
		}
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

	err := c.parseFlags(ctx)

	if err != nil {
		return err
	}

	// If there are flags, extra args have been checked for so default
	// is true. If there are no flags associated with command, make sure
	// the actionable command is the last arg.
	noExtraArgs := true
	if !c.hasFlags() {
		noExtraArgs = ctx.Index == len(ctx.Args)
	}
	if c.isValid() {
		if c.Action != nil && noExtraArgs {
			return c.Action(c, ctx)
		}

		if c.SubCommands != nil && ctx.Index < len(ctx.Args) {
			sc := c.SubCommands[ctx.Args[ctx.Index]]

			if sc == nil {
				return errors.New("Invalid command '" + ctx.Args[ctx.Index] + "'")
			}

			return sc.Execute(ctx)
		}
	}
	c.printUsage()
	// Extra input after the command is an error
	if c.isValid() && !noExtraArgs && !helpFlags[ctx.Args[ctx.Index]] {
		fmt.Print("\n-----\n")
		return fmt.Errorf("Unrecognized input: %v\n", ctx.Args[ctx.Index])
	}

	return nil
}

func (c *Command) parseFlags(ctx *Context) error {
	if c.flags != nil {
		c.flags.Parse(ctx.Args[ctx.Index:])
		ctx.Index += c.flags.NFlag()
	}

	if c.SubCommands == nil && c.hasFlags() && len(c.flags.Args()) > 0 {
		c.printUsage()
		fmt.Print("\n-----\n")
		return fmt.Errorf("Unrecognized input: %s\n", c.flags.Args())
	}
	return nil
}

// isInList looks for a string in a list of strings.
func isInList(item string, list []string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

// FIXME: allow non-echoing for passwords
func promptStdIn(prompt string) (string, error) {
	bio := bufio.NewReader(os.Stdin)
	fmt.Printf(prompt)
	line, _, err := bio.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func getUrlForResource(apiPath string, id uint64) string {
	return fmt.Sprintf("%s/%d", apiPath, id)
}
