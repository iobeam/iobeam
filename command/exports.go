package command

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

type exportData struct {
	projectId uint64
	deviceId  string
	series    string

	limit       uint64
	from        uint64
	to          uint64
	lessThan    int64
	greaterThan int64
	equal       string
}

func (e *exportData) IsValid() bool {
	pidOk := e.projectId > 0
	limitOk := e.limit > 0
	rangeOk := e.from <= e.to
	valRangeOk := e.greaterThan <= e.greaterThan
	equalOk := len(e.equal) == 0
	if !equalOk {
		_, err := strconv.ParseInt(e.equal, 0, 64)
		equalOk = err == nil
	}
	return pidOk && limitOk && rangeOk && valRangeOk && equalOk
}

// NewExportCommand returns the base 'export' command.
func NewExportCommand(ctx *Context) *Command {
	e := new(exportData)
	pid := ctx.Profile.ActiveProject

	cmd := &Command{
		Name:    "query",
		ApiPath: "/v1/exports",
		Usage:   "Get data for projects, devices, and series",
		Data:    e,
		Action:  getExport,
	}

	flags := cmd.NewFlagSet("iobeam query")
	maxTime := uint64((time.Now().UnixNano() / int64(time.Millisecond)) + (1000 * 60 * 60 * 24))
	flags.Uint64Var(&e.projectId, "projectId", pid, "Project ID (if omitted, defaults to active project)")
	flags.StringVar(&e.deviceId, "deviceId", "", "Device ID")
	flags.StringVar(&e.series, "series", "", "Series name")

	flags.Uint64Var(&e.limit, "limit", 10, "Max number of results")
	flags.Uint64Var(&e.from, "from", 0, "Min timestamp (unix time in milliseconds)")
	flags.Uint64Var(&e.to, "to", maxTime, "Max timestamp (unix time in milliseconds, default is now + a day)")
	flags.Int64Var(&e.lessThan, "lessThan", math.MaxInt64, "Max value for datapoints")
	flags.Int64Var(&e.greaterThan, "greaterThan", math.MinInt64, "Min value for datapoints")
	flags.StringVar(&e.equal, "equalTo", "", "Datapoints with this value")
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

	req := ctx.Client.Get(reqPath).Expect(200).ProjectToken(ctx.Profile, e.projectId)
	req = req.
		ParamUint64("limit", e.limit).
		ParamUint64("from", e.from).
		ParamUint64("to", e.to).
		ParamInt64("less_than", e.lessThan).
		ParamInt64("greater_than", e.greaterThan)
	if len(e.equal) > 0 {
		temp, _ := strconv.ParseInt(e.equal, 0, 64)
		fmt.Println(temp)
		req = req.ParamInt64("equals", temp)
	}

	x := make(map[string]interface{})
	_, err := req.ResponseBody(&x).
		ResponseBodyHandler(func(token interface{}) error {
		fmt.Println("Results: ")
		output, err := json.MarshalIndent(token, "", "  ")
		fmt.Println(string(output))
		return err
	}).Execute()

	return err
}
