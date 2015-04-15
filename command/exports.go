package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
)

type exportData struct {
	projectId uint64
	deviceId  string
	series    string
	limit     uint64
}

func (e *exportData) IsValid() bool {
	return e.projectId > 0 && e.limit > 0
}

// NewExportCommand returns the base 'export' command.
func NewExportCommand(ctx *Context) *Command {
	e := new(exportData)

	cmd := &Command{
		Name:    "query",
		ApiPath: "/v1/exports",
		Usage:   "Get data for projects, devices, and series",
		Data:    e,
		Flags:   flag.NewFlagSet("exports", flag.ExitOnError),
		Action:  getExport,
	}

	pid := ctx.Profile.ActiveProject
	cmd.Flags.Uint64Var(&e.projectId, "projectId", pid, "Project ID (REQUIRED - defaults to active project)")
	cmd.Flags.StringVar(&e.deviceId, "deviceId", "", "Device ID")
	cmd.Flags.StringVar(&e.series, "series", "", "Series name")
	cmd.Flags.Uint64Var(&e.limit, "limit", 10, "Max number of results")

	return cmd
}

// getExport fetches the requested data from the iobeam Cloud based on
// the provided projectID, deviceID, and series name.
func getExport(c *Command, ctx *Context) error {
	e := c.Data.(*exportData)

	reqPath := c.ApiPath + "/" + strconv.FormatUint(e.projectId, 10)
	device := "all"
	if len(e.deviceId) > 0 {
		device = e.deviceId
	}
	reqPath += "/" + device
	if len(e.series) > 0 {
		reqPath += "/" + e.series
	}

	x := make(map[string]interface{})
	_, err := ctx.Client.
		Get(reqPath).
		Expect(200).
		ProjectToken(ctx.Profile, e.projectId).
		ParamUint64("limit", e.limit).
		ResponseBody(&x).
		ResponseBodyHandler(func(token interface{}) error {
		fmt.Println("Results: ")
		output, err := json.MarshalIndent(token, "", "  ")
		fmt.Println(string(output))
		return err
	}).Execute()

	return err
}
