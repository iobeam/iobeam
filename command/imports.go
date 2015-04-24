package command

import (
	"flag"
	"fmt"
	"time"
)

type importData struct {
	projectId uint64
	deviceId  string
	series    string
	timestamp int64
	value     float64
}

func (d *importData) IsValid() bool {
	return d.projectId > 0 && len(d.deviceId) > 0 && len(d.series) > 0 && d.timestamp > 0
}

// NewImportCommand returns the base 'import' command.
func NewImportCommand(ctx *Context) *Command {
	d := new(importData)

	cmd := &Command{
		Name:    "import",
		ApiPath: "/v1/imports",
		Usage:   "Add data to a project",
		Data:    d,
		Flags:   flag.NewFlagSet("iobeam import", flag.ExitOnError),
		Action:  sendImport,
	}

	pid := ctx.Profile.ActiveProject
	now := time.Now().UnixNano() / int64(time.Millisecond)
	cmd.Flags.Uint64Var(&d.projectId, "projectId", pid, "Project ID (if omitted, defaults to active project)")
	cmd.Flags.StringVar(&d.deviceId, "deviceId", "", "Device ID (REQUIRED)")
	cmd.Flags.StringVar(&d.series, "series", "", "Series name (REQUIRED)")
	cmd.Flags.Int64Var(&d.timestamp, "time", now, "Timestamp, in milliseconds, of the data (if omitted, defaults to current time)")
	cmd.Flags.Float64Var(&d.value, "value", 0.0, "Data value (REQUIRED)")

	return cmd
}

type dataPoint struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

type sourceObj struct {
	Name string      `json:"name"`
	Data []dataPoint `json:"data"`
}

type importObj struct {
	ProjectId uint64      `json:"project_id"`
	DeviceId  string      `json:"device_id"`
	Sources   []sourceObj `json:"sources"`
}

func sendImport(c *Command, ctx *Context) error {
	d := c.Data.(*importData)

	dp := dataPoint{
		Time:  d.timestamp,
		Value: d.value,
	}
	source1 := sourceObj{
		Name: d.series,
		Data: []dataPoint{dp},
	}
	obj := importObj{
		ProjectId: d.projectId,
		DeviceId:  d.deviceId,
		Sources:   []sourceObj{source1},
	}

	_, err := ctx.Client.
		Post(c.ApiPath).
		Expect(200).
		ProjectToken(ctx.Profile, d.projectId).
		Body(obj).
		Execute()

	if err == nil {
		fmt.Println("Data successfully imported.")
	}

	return err
}
