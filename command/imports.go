package command

import (
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
	return d.projectId > 0 && len(d.deviceId) > 0 && len(d.series) > 0 && d.timestamp >= 0
}

// NewImportCommand returns the base 'import' command.
func NewImportCommand(ctx *Context) *Command {
	d := new(importData)
	pid := ctx.Profile.ActiveProject
	now := time.Now().UnixNano() / int64(time.Millisecond)

	cmd := &Command{
		Name:    "import",
		ApiPath: "/v1/imports",
		Usage:   "Add new data points.",
		Data:    d,
		Action:  sendImport,
	}

	flags := cmd.NewFlagSet("iobeam import")
	flags.Uint64Var(&d.projectId, "projectId", pid, "Project ID (if omitted, defaults to active project)")
	flags.StringVar(&d.deviceId, "deviceId", "", "Device ID (REQUIRED)")
	flags.StringVar(&d.series, "series", "", "Series name (REQUIRED)")
	flags.Int64Var(&d.timestamp, "time", now, "Timestamp, in milliseconds, of the data (if omitted, defaults to current time)")
	flags.Float64Var(&d.value, "value", 0.0, "Data value (REQUIRED)")

	return cmd
}

type dataPoint struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

type sourceObj struct {
	Fields []string        `json:"fields"`
	Data   [][]interface{} `json:"data"`
}

type importObj struct {
	ProjectId uint64    `json:"project_id"`
	DeviceId  string    `json:"device_id"`
	Sources   sourceObj `json:"sources"`
}

func sendImport(c *Command, ctx *Context) error {
	d := c.Data.(*importData)

	fields := []string{"time", d.series}
	row := []interface{}{d.timestamp, d.value}

	source := sourceObj{
		Fields: fields,
		Data:   [][]interface{}{row},
	}
	obj := importObj{
		ProjectId: d.projectId,
		DeviceId:  d.deviceId,
		Sources:   source,
	}

	_, err := ctx.Client.
		Post(c.ApiPath).
		Expect(200).
		ProjectToken(ctx.Profile, d.projectId).
		Param("fmt", "table").
		Body(obj).
		Execute()

	if err == nil {
		fmt.Println("Data successfully imported.")
	}

	return err
}
