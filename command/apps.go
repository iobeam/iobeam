package command

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/iobeam/iobeam/client"
)

const (
	baseApiPathApp = "/v1/apps"
	flagSetApp     = "iobeam app"
	cmdGet         = "get"
	cmdDelete      = "delete"
	cmdLaunch      = "deploy"
	cmdList        = "list"
	cmdStart       = "start"
	cmdStop        = "stop"

	appStatusRunning = "RUNNING"
	appStatusStopped = "STOPPED"
)

func (c *Command) newFlagSetApp(cmd string) *flag.FlagSet {
	return c.NewFlagSet(flagSetApp + " " + cmd)
}

// bundle describes an app's file bundle
type bundle struct {
	Checksum checksum `json:"checksum"`
	Type     string   `json:"type"`
	URI      string   `json:"uri"`
}

func (b *bundle) Print() {
	fmt.Printf("URI     : %s\n", b.URI)
	fmt.Printf("Type    : %s\n", b.Type)
	b.Checksum.Print()
	fmt.Println()
}

// appData defines the JSON resource of an iobeam app
type appData struct {
	AppId           uint64 `json:"app_id,omitempty"`
	AppName         string `json:"app_name"`
	ProjectId       uint64 `json:"project_id"`
	Bundle          bundle `json:"bundle"`
	Created         string `json:"created,omitempty"`
	LastMod         string `json:"last_modified,omitempty"`
	RequestedStatus string `json:"requested_status,omitempty"`
	CurrentStatus   string `json:"current_status,omitempty"`
}

func (i *appData) Print() {
	fmt.Printf("App ID  : %d\n", i.AppId)
	fmt.Printf("App Name: %s\n", i.AppName)
	fmt.Printf("Created : %s\n", i.Created)
	if len(i.RequestedStatus) > 0 {
		fmt.Printf("Requested Status: %s\n", i.RequestedStatus)
	}
	if len(i.CurrentStatus) > 0 {
		fmt.Printf("Current Status  : %s\n", i.CurrentStatus)
	}
	fmt.Println()
	fmt.Println("BUNDLE INFO")
	i.Bundle.Print()
}

// NewAppsCommand returns the base 'app' command.
func NewAppsCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "app",
		Usage: "Commands for managing apps.",
		SubCommands: Mux{
			cmdLaunch: newLaunchAppCmd(ctx),
			cmdDelete: newDeleteAppCmd(ctx),
			cmdGet:    newGetAppCmd(ctx),
			cmdList:   newListAppsCmd(ctx),
			cmdStart:  newStartAppCmd(ctx),
			cmdStop:   newStopAppCmd(ctx),
		},
	}
	cmd.NewFlagSet(flagSetApp)

	return cmd
}

// launchAppArgs are the arguments for the 'launch' subcommand
type launchAppArgs struct {
	uploadFileArgs
	name string
}

func (a *launchAppArgs) IsValid() bool {
	return a.uploadFileArgs.IsValid() && len(a.name) > 0
}

func newLaunchAppCmd(ctx *Context) *Command {
	args := new(launchAppArgs)

	cmd := &Command{
		Name:    cmdLaunch,
		ApiPath: baseApiPathApp,
		Usage:   "Launch a Spark app on iobeam.",
		Data:    args,
		Action:  launchApp,
	}
	flags := cmd.newFlagSetApp(cmdLaunch)
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID (defaults to active project).")
	flags.StringVar(&args.name, "name", "", "Name of the app. (REQUIRED)")
	flags.StringVar(&args.path, "path", "", "Path to app to upload. (REQUIRED)")

	return cmd
}

func launchApp(c *Command, ctx *Context) error {
	args := c.Data.(*launchAppArgs)
	digest, err := _uploadFile(ctx, &args.uploadFileArgs)
	if err != nil {
		return err
	}

	data := &appData{
		AppName:         args.name,
		ProjectId:       args.projectId,
		RequestedStatus: appStatusRunning,
		Bundle: bundle{
			Type: "JAR",
			URI:  "file://" + filepath.Base(args.path),
			Checksum: checksum{
				Sum:       digest,
				Algorithm: "SHA256",
			},
		},
	}

	_, err = ctx.Client.
		Post(c.ApiPath).
		Expect(201).
		ProjectToken(ctx.Profile, args.projectId).
		Body(data).
		ResponseBody(data).
		ResponseBodyHandler(func(body interface{}) error {

		app := body.(*appData)
		fmt.Println("New app created.")
		fmt.Printf("App ID: %v\n", app.AppId)
		fmt.Printf("App Name: %v\n", app.AppName)
		fmt.Println()

		return nil
	}).Execute()

	return err
}

// baseAppArgs are the very basic arguments for subcommands
type baseAppArgs struct {
	projectId uint64
	name      string
	id        uint64
}

func (a *baseAppArgs) IsValid() bool {
	return a.projectId > 0 && (len(a.name) > 0 || a.id > 0)
}

// getAppArgs are the arguments for the 'get' subcommand
type getAppArgs struct {
	baseAppArgs
}

func (a *getAppArgs) IsValid() bool {
	return a.baseAppArgs.IsValid()
}

func newGetAppCmd(ctx *Context) *Command {
	args := new(getAppArgs)

	cmd := &Command{
		Name:    cmdGet,
		ApiPath: baseApiPathApp,
		Usage:   "Get app information.",
		Data:    args,
		Action:  getApp,
	}
	flags := cmd.newFlagSetApp(cmdGet)
	flags.Uint64Var(&args.id, "id", 0, "App ID to get (this or -name is required)")
	flags.StringVar(&args.name, "name", "", "App name to get (this or -id is required)")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID to get from (defaults to active project)")

	return cmd
}

func getApp(c *Command, ctx *Context) error {
	args := c.Data.(*getAppArgs)
	app, err := _getApp(ctx, &args.baseAppArgs)
	if err == nil {
		app.Print()
	}
	return err
}

func _getApp(ctx *Context, args *baseAppArgs) (*appData, error) {
	var req *client.Request
	if args.id > 0 {
		req = ctx.Client.Get(baseApiPathApp + "/" + strconv.FormatUint(args.id, 10))
	} else {
		req = ctx.Client.Get(baseApiPathApp).Param("name", args.name)
	}

	app := new(appData)
	_, err := req.Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(app).
		ResponseBodyHandler(func(body interface{}) error {
		return nil
	}).Execute()

	return app, err
}

// deleteAppArgs are the arguments for the 'delete' subcommand
type deleteAppArgs struct {
	baseAppArgs
}

func (a *deleteAppArgs) IsValid() bool {
	return a.baseAppArgs.IsValid()
}

func newDeleteAppCmd(ctx *Context) *Command {
	args := new(deleteAppArgs)

	cmd := &Command{
		Name:    cmdDelete,
		ApiPath: baseApiPathApp,
		Usage:   "Delete an app.",
		Data:    args,
		Action:  deleteApp,
	}
	flags := cmd.newFlagSetApp(cmdDelete)
	flags.Uint64Var(&args.id, "id", 0, "App ID to delete (REQUIRED)")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID to delete from (defaults to active project)")

	return cmd
}

func deleteApp(c *Command, ctx *Context) error {
	args := c.Data.(*deleteAppArgs)
	req := ctx.Client.Delete(baseApiPathApp + "/" + strconv.FormatUint(args.id, 10))

	_, err := req.Expect(204).
		ProjectToken(ctx.Profile, args.projectId).Execute()

	if err == nil {
		fmt.Printf("App %d sucessfully deleted.\n", args.id)
	}

	return err
}

// listAppsArgs are the arguments for the 'list' sucommand
type listAppsArgs struct {
	projectId uint64
}

func (a *listAppsArgs) IsValid() bool {
	return a.projectId > 0
}

func newListAppsCmd(ctx *Context) *Command {
	args := new(listAppsArgs)

	cmd := &Command{
		Name:    cmdList,
		ApiPath: baseApiPathApp,
		Usage:   "List apps for a project.",
		Data:    args,
		Action:  listApps,
	}
	flags := cmd.newFlagSetApp(cmdList)
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to list from (defaults to active project).")

	return cmd
}

func listApps(c *Command, ctx *Context) error {
	type listResult struct {
		Apps []appData `json:"apps"`
	}
	args := c.Data.(*listAppsArgs)

	_, err := ctx.Client.
		Get(c.ApiPath).
		Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(new(listResult)).
		ResponseBodyHandler(func(body interface{}) error {
		list := body.(*listResult)
		if len(list.Apps) > 0 {
			spacer := ""
			for _, info := range list.Apps {
				fmt.Printf(spacer)
				info.Print()
				spacer = "----------\n"
			}
		} else {
			fmt.Printf("No apps found for project %d.\n", args.projectId)
		}

		return nil
	}).Execute()

	return err
}

func newStartAppCmd(ctx *Context) *Command {
	args := new(baseAppArgs)

	cmd := &Command{
		Name:    cmdStart,
		ApiPath: baseApiPathApp,
		Usage:   "Start an app by name/id.",
		Data:    args,
		Action:  startApp,
	}
	flags := cmd.newFlagSetApp(cmdStart)
	flags.Uint64Var(&args.id, "id", 0, "App ID to start (REQUIRED).")
	//flags.StringVar(&args.name, "name", "", "App name to start (this or -id is required)")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID of app (defaults to active project)")

	return cmd
}

func startApp(c *Command, ctx *Context) error {
	args := c.Data.(*baseAppArgs)
	return _updateAppStatus(ctx, args, appStatusRunning)
}

func newStopAppCmd(ctx *Context) *Command {
	args := new(baseAppArgs)

	cmd := &Command{
		Name:    cmdStop,
		ApiPath: baseApiPathApp,
		Usage:   "Stop an app by name/id.",
		Data:    args,
		Action:  stopApp,
	}
	flags := cmd.newFlagSetApp(cmdStart)
	flags.Uint64Var(&args.id, "id", 0, "App ID to stop (REQUIRED).")
	//flags.StringVar(&args.name, "name", "", "App name to stop (this or -id is required)")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID of app (defaults to active project)")

	return cmd
}

func stopApp(c *Command, ctx *Context) error {
	args := c.Data.(*baseAppArgs)
	return _updateAppStatus(ctx, args, appStatusStopped)
}

func _updateAppStatus(ctx *Context, args *baseAppArgs, status string) error {
	app, err := _getApp(ctx, args)
	if err != nil {
		return err
	}

	var req *client.Request
	if args.id > 0 {
		req = ctx.Client.Patch(baseApiPathApp + "/" + strconv.FormatUint(args.id, 10))
	} else {
		req = ctx.Client.Patch(baseApiPathApp).Param("name", args.name)
	}

	app.RequestedStatus = status
	rsp, err := req.Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		Body(app).
		Execute()

	if err != nil && rsp.Http().StatusCode == 204 {
		return nil
	}

	return err
}
