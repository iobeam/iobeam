package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"iobeam.com/iobeam/client"
	"strconv"
)

type exportData struct {
	projectId uint64
	deviceId  string
	series    string
	limit     uint64
}

func (e *exportData) IsValid() bool {
	// TODO: Add conditional to check that device ID must be set if series is.
	return e.projectId > 0 && e.limit > 0
}

// NewExportCommand returns the base 'export' command.
func NewExportCommand() *Command {

	e := new(exportData)

	cmd := &Command{
		Name:    "export",
		ApiPath: "/v1/exports",
		Usage:   "Get data for projects, devices, and series",
		Data:    e,
		Flags:   flag.NewFlagSet("exports", flag.ExitOnError),
		Action:  getExport,
	}

	cmd.Flags.Uint64Var(&e.projectId, "projectId", 0, "Project ID (REQUIRED)")
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
	if len(e.deviceId) > 0 {
		reqPath += "/" + e.deviceId

		if len(e.series) > 0 {
			reqPath += "/" + e.series
		}
	}

	token, err := client.ReadProjToken(e.projectId)
	if err != nil {
		fmt.Printf("Missing token for project %d.\n", e.projectId)
		return err
	}

	x := make(map[string]interface{})
	_, err = ctx.Client.
		Get(reqPath).
		Expect(200).
		Token(token).
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
