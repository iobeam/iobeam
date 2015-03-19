package command

import (
	"os"
	"fmt"
	"flag"
	"errors"
	"beam.io/beam/client"
)

// Interface fo data that is posted to API, and generated from command-line input.
type Data interface {
	IsValid() bool
}

type Context struct {
	Cmd *Command
	Client *client.Client
	Index int
	Args []string
}

type Mux map[string]*Command

type CommandAction func(*Command, *Context) (error)

type Command struct {
	Name string
	Usage string
	ApiPath string
	Flags *flag.FlagSet
	SubCommands Mux
	Data Data
	Action CommandAction
}

func (c *Command) PrintUsage() {

	if c.SubCommands != nil {
		fmt.Fprintf(os.Stderr, "Usage: %s COMMAND [FLAGS]\n\n", c.Name)
		fmt.Fprintf(os.Stderr, "%s\n", c.Usage)

		fmt.Fprint(os.Stderr, "\nAvailable Commands:\n")

		for _, v := range(c.SubCommands) {
			fmt.Fprintf(os.Stderr, "  %-20s :: %s\n",
				v.Name, v.Usage)
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

func (c *Command) IsValid() bool {
	if c.Data == nil {
		return true
	}

	return c.Data.IsValid()
}
	

func (c *Command) Execute(ctx *Context) (error) {
	
	ctx.Index++
	
	c.ParseFlags(ctx)

	if c.IsValid() {	
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

func (c *Command) ParseFlags(ctx *Context) {
	if c.Flags != nil {
		//fmt.Println("Parsing flags for command", c.Name, "arr", ctx.Args[ctx.Index:])
		c.Flags.Parse(ctx.Args[ctx.Index:])
		ctx.Index += c.Flags.NFlag()
	}
}
