package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type importData struct {
	projectId uint64
	deviceId  string
	namespace string
	series    string
	timestamp int64
	value     string
}

func (d *importData) IsValid() bool {
	return d.projectId > 0 && len(d.deviceId) > 0 && len(d.series) > 0 && d.timestamp >= 0 && IsValidType(d.value)
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
	flags.StringVar(&d.namespace, "namespace", "input", "Namespace to write to (Defaults to 'input')")
	flags.StringVar(&d.series, "series", "", "Series name (REQUIRED)")
	flags.Int64Var(&d.timestamp, "time", now, "Timestamp, in milliseconds, of the data (if omitted, defaults to current time)")
	flags.StringVar(&d.value, "value", "", "Data value (REQUIRED)")

	return cmd
}

type dataObj struct {
	Fields []string        `json:"fields"`
	Data   [][]interface{} `json:"data"`
}

type importObj struct {
	ProjectId uint64  `json:"project_id"`
	DeviceId  string  `json:"device_id"`
	Data      dataObj `json:"data"`
	Namespace string  `json:"namespace"`
}

func createDatapoint(timestamp int64, value string) ([]interface{}, error) {

	if IsValidDouble(value) {
		doubleValue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return []interface{}{timestamp, doubleValue}, nil
		}
		return nil, err
	}

	if IsValidLong(value) {
		longValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return []interface{}{timestamp, longValue}, nil
		}
		return nil, err
	}

	if IsValidString(value) {
		// remove the quotes
		str := strings.TrimSuffix(strings.TrimPrefix(value, "\""), "\"")

		return []interface{}{timestamp, str}, nil
	}

	return nil, errors.New("Failed to determine type of series value.")

}

func sendImport(c *Command, ctx *Context) error {
	d := c.Data.(*importData)

	fields := []string{"time", d.series}

	row, err := createDatapoint(d.timestamp, d.value)

	if err != nil {
		return err
	}

	data := dataObj{
		Fields: fields,
		Data:   [][]interface{}{row},
	}
	obj := importObj{
		ProjectId: d.projectId,
		DeviceId:  d.deviceId,
		Namespace: d.namespace,
		Data:      data,
	}

	_, err = ctx.Client.
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
