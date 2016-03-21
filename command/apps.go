package command

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/iobeam/iobeam/client"
)

const (
	appStatusRunning = "RUNNING"
	appStatusStopped = "STOPPED"
	appStatusError   = "ERROR"

	keyApp = "app"

	maxStatusTries = 10
	backOffAmt     = 3
	backOffMax     = 10
)

func init() {
	flagSetNames[keyApp] = "iobeam app"
	baseApiPath[keyApp] = "/v1/apps"
}

func (c *Command) newFlagSetApp() *flag.FlagSet {
	return c.NewFlagSet(flagSetNames[keyApp] + " " + c.Name)
}

func getUrlforAppId(id uint64) string {
	return getUrlForResource(baseApiPath[keyApp], id)
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
	Error           string `json:"error,omitempty"`
}

func (i *appData) Print() {
	fmt.Printf("App ID  : %d\n", i.AppId)
	fmt.Printf("App Name: %s\n", i.AppName)
	fmt.Printf("Created : %s\n", i.Created)
	if len(i.RequestedStatus) > 0 {
		fmt.Printf("Requested Status: %s\n", i.RequestedStatus)
	}
	if len(i.CurrentStatus) > 0 {
		fmt.Printf("Current Status  : %s", i.CurrentStatus)
		if len(i.Error) > 0 {
			fmt.Printf(" (%s)\n", i.Error)
		} else {
			fmt.Printf("\n")
		}
	}
	fmt.Println()
	fmt.Println("BUNDLE INFO")
	i.Bundle.Print()
}

// NewAppsCommand returns the base 'app' command.
func NewAppsCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  keyApp,
		Usage: "Commands for managing apps.",
		SubCommands: Mux{
			"create": newLaunchAppCmd(ctx),
			"delete": newDeleteAppCmd(ctx),
			"get":    newGetAppCmd(ctx),
			"list":   newListAppsCmd(ctx),
			"start":  newStartAppCmd(ctx),
			"stop":   newStopAppCmd(ctx),
			"update": newUpdateAppCmd(ctx),
		},
	}
	cmd.NewFlagSet(flagSetNames[keyApp])

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
		Name:    "create",
		ApiPath: baseApiPath[keyApp],
		Usage:   "Create (and launch) a Spark app.",
		Data:    args,
		Action:  launchApp,
	}
	flags := cmd.newFlagSetApp()
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

// updateAppArgs are the arguments for the 'update' subcommand
type updateAppArgs struct {
	launchAppArgs
	id uint64
}

func (a *updateAppArgs) IsValid() bool {
	return a.id > 0 && (len(a.name) > 0 || len(a.path) > 0)
}

func newUpdateAppCmd(ctx *Context) *Command {
	args := new(updateAppArgs)

	cmd := &Command{
		Name: "update",
		// ApiPath determined by flags
		Usage:  "Update an app, including replacing the JAR.",
		Data:   args,
		Action: updateApp,
	}
	flags := cmd.newFlagSetApp()
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID (defaults to active project).")
	flags.Uint64Var(&args.id, "id", 0, "App ID to update. (REQUIRED)")
	flags.StringVar(&args.name, "name", "", "Name of the app.")
	flags.StringVar(&args.path, "path", "", "Path to app to upload.")

	return cmd
}

func updateApp(c *Command, ctx *Context) error {
	args := c.Data.(*updateAppArgs)

	// Get app info to do PUT
	app, err := _getApp(ctx, &baseAppArgs{
		id:        args.id,
		projectId: args.projectId,
	})
	if err != nil {
		return err
	}

	if len(args.name) > 0 {
		app.AppName = args.name
	}

	if len(args.path) > 0 {
		digest, err := _uploadFile(ctx, &args.uploadFileArgs)
		if err != nil {
			return err
		}
		app.Bundle = bundle{
			Type: "JAR",
			URI:  "file://" + filepath.Base(args.path),
			Checksum: checksum{
				Sum:       digest,
				Algorithm: "SHA256",
			},
		}
	}

	rsp, err := ctx.Client.
		Put(getUrlforAppId(args.id)).
		Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		Body(app).
		Execute()

	if err == nil {
		fmt.Println("App successfully updated.")
	} else if rsp.Http().StatusCode == 204 {
		fmt.Println("App not modified.")
		return nil
	}

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
		Name: "get",
		// ApiPath determined by flags
		Usage:  "Get app information.",
		Data:   args,
		Action: getApp,
	}
	flags := cmd.newFlagSetApp()
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
		req = ctx.Client.Get(getUrlforAppId(args.id))
	} else {
		req = ctx.Client.Get(baseApiPath[keyApp]).Param("name", args.name)
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
		Name: "delete",
		// ApiPath determined by flags
		Usage:  "Delete an app.",
		Data:   args,
		Action: deleteApp,
	}
	flags := cmd.newFlagSetApp()
	flags.Uint64Var(&args.id, "id", 0, "App ID to delete (REQUIRED)")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID to delete from (defaults to active project)")

	return cmd
}

func deleteApp(c *Command, ctx *Context) error {
	args := c.Data.(*deleteAppArgs)
	req := ctx.Client.Delete(getUrlforAppId(args.id))

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
		Name:    "list",
		ApiPath: baseApiPath[keyApp],
		Usage:   "List apps for a project.",
		Data:    args,
		Action:  listApps,
	}
	flags := cmd.newFlagSetApp()
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
					fmt.Println()
					spacer = "----------\n\n"
				}
			} else {
				fmt.Printf("No apps found for project %d.\n", args.projectId)
			}

			return nil
		}).Execute()

	return err
}

type appStartOrStopArgs struct {
	baseAppArgs
	isStart bool
}

func (a *appStartOrStopArgs) IsValid() bool {
	return a.baseAppArgs.IsValid()
}

func newStartOrStopAppCmd(ctx *Context, isStart bool) *Command {
	args := &appStartOrStopArgs{isStart: isStart}
	var cmdStr string
	var desc string
	if isStart {
		cmdStr = "start"
		desc = "Start an app by ID."
	} else {
		cmdStr = "stop"
		desc = "Stop an app by ID."
	}

	cmd := &Command{
		Name: cmdStr,
		// ApiPath determined by flags
		Usage:  desc,
		Data:   args,
		Action: updateAppStatus,
	}
	flags := cmd.newFlagSetApp()
	flags.Uint64Var(&args.id, "id", 0, "App ID (REQUIRED).")
	//flags.StringVar(&args.name, "name", "", "App name (this or -id is required)")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID of app (defaults to active project)")

	return cmd
}

func newStartAppCmd(ctx *Context) *Command {
	return newStartOrStopAppCmd(ctx, true /* isStart */)
}

func newStopAppCmd(ctx *Context) *Command {
	return newStartOrStopAppCmd(ctx, false /* isStart */)
}

func updateAppStatus(c *Command, ctx *Context) error {
	args := c.Data.(*appStartOrStopArgs)
	app, err := _getApp(ctx, &args.baseAppArgs)
	if err != nil {
		return err
	}

	var req *client.Request
	if args.id > 0 {
		req = ctx.Client.Put(getUrlforAppId(args.id))
	} else {
		req = ctx.Client.Put(baseApiPath[keyApp]).Param("name", args.name)
	}

	var wantedStatus string
	if args.isStart {
		wantedStatus = appStatusRunning
	} else {
		wantedStatus = appStatusStopped
	}
	app.RequestedStatus = wantedStatus
	rsp, err := req.
		Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		Body(app).
		Execute()

	if err != nil && rsp.Http().StatusCode == 204 {
		fmt.Printf("Requested status is already %s\n", wantedStatus)
		return nil
	} else if err != nil {
		return err
	}

	fmt.Printf("Requested status: %s. Waiting for current status to change.\n", wantedStatus)
	tries := 0
	sleep := 0
	for true {
		fmt.Printf("Checking app status...")
		time.Sleep(time.Duration(sleep) * time.Second)

		app, err := _getApp(ctx, &args.baseAppArgs)
		if err != nil {
			return fmt.Errorf("Error while waiting for status change: %v\n", err)
		}
		fmt.Printf("%s\n", app.CurrentStatus)

		if app.CurrentStatus == wantedStatus {
			fmt.Printf("Success!\n")
			break
		} else if app.CurrentStatus == appStatusError {
			fmt.Printf("Unsuccessful, app finished in error state.\n")
			break
		}

		sleep += backOffAmt
		if sleep > backOffMax {
			sleep = backOffMax
		}

		tries++
		if tries > maxStatusTries {
			fmt.Printf("Timed out waiting for app status to change.\n")
			break
		}
	}

	return nil
}
