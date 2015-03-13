package command

import (
	"os"
	"fmt"
	"flag"
	"errors"
	"watchvast.com/cerebctl/client"
)

type Context struct {
	Cmd *Command
	Client *client.Client
	Index int
	Args []string
}

type Mux map[string]*Command

type Command struct {
	Name string
	Usage string
	ApiPath string
	Flags *flag.FlagSet
	SubCommands Mux
	Data Data
	Action func(*Command, *Context) (error)
}

func (c *Command) PrintUsage() {

	if c.SubCommands != nil {
		fmt.Fprintf(os.Stderr, "Usage: %s COMMAND [FLAGS]\n", c.Name)

		fmt.Fprint(os.Stderr, "\nAvailable Commands:\n")

		for _, v := range(c.SubCommands) {
			fmt.Fprintf(os.Stderr, "  %-20s :: %s\n",
				v.Name, v.Usage)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Usage: %s [FLAGS]\n", c.Name)
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
	}
}
