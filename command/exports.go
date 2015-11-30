package command

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	opSum   = "sum"
	opCount = "count"
	opMin   = "min"
	opMax   = "max"
	opMean  = "mean"
)

const (
	timeFmtSec    = "sec"
	timeFmtMsec   = "msec"
	timeFmtUsec   = "usec"
	timeFmtStruct = "timeval"
)

const (
	outputJson = "json"
	outputCsv  = "csv"
)

const max_duration_str = "24h"

var ops = []string{opSum, opCount, opMin, opMax, opMean}
var timeFmts = []string{timeFmtSec, timeFmtMsec, timeFmtUsec, timeFmtStruct}
var outputs = []string{outputJson, outputCsv}

type exportData struct {
	projectId uint64
	deviceId  string
	series    string
	timeFmt   string
	output    string

	limit       uint64
	from        uint64
	to          uint64
	lessThan    int64
	greaterThan int64
	equal       string

	operator string
	groupBy  string
}

func isInList(item string, list []string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
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

	opOk := len(e.operator) == 0 || isInList(e.operator, ops)

	groupOk := len(e.groupBy) == 0
	if !groupOk {
		_, err := time.ParseDuration(e.groupBy)
		groupOk = err == nil && len(e.operator) > 0
	}

	timeOk := isInList(e.timeFmt, timeFmts)
	outputOk := isInList(e.output, outputs)

	return pidOk && limitOk && rangeOk && valRangeOk && equalOk && opOk && groupOk && timeOk && outputOk
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
	max_duration, _ := time.ParseDuration(max_duration_str)
	maxTime := time.Now().Add(max_duration).UnixNano() / int64(time.Millisecond)
	flags.Uint64Var(&e.projectId, "projectId", pid, "Project ID (if omitted, defaults to active project)")
	flags.StringVar(&e.deviceId, "deviceId", "", "Device ID")
	flags.StringVar(&e.series, "series", "", "Series name")

	flags.Uint64Var(&e.limit, "limit", 10, "Max number of results")
	flags.Uint64Var(&e.from, "from", 0, "Min timestamp (unix time in milliseconds)")
	flags.Uint64Var(&e.to, "to", uint64(maxTime), "Max timestamp (unix time in milliseconds, default is now + a day)")
	flags.Int64Var(&e.lessThan, "lessThan", math.MaxInt64, "Max value for datapoints")
	flags.Int64Var(&e.greaterThan, "greaterThan", math.MinInt64, "Min value for datapoints")
	flags.StringVar(&e.equal, "equalTo", "", "Datapoints with this value")
	flags.StringVar(&e.operator, "operator", "", "Aggregation function to apply to datapoints: "+strings.Join(ops, ", "))
	flags.StringVar(&e.groupBy, "groupBy", "", "Group data by [number][period], where the time period can be ms, s, m, or h (e.g., 30s, 15m, 6h). Requires a valid operator.")
	flags.StringVar(&e.timeFmt, "timeFmt", "msec", "Time unit to display timestamps: "+strings.Join(timeFmts, ", "))
	flags.StringVar(&e.output, "output", "json", "Output format of the results. Valid outputs: json, csv")
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

	req := ctx.Client.Get(reqPath).Expect(200).
		ProjectToken(ctx.Profile, e.projectId).
		ParamUint64("limit", e.limit).
		Param("timefmt", e.timeFmt).
		Param("output", e.output)

	// Only add params if actually set / necessary, i.e.:
	// - "to" is less than current time
	// - "from" is something other than 0
	// - "lessThan" is something other than MAX INT
	// - "greaterThan" is something other than MIN INT
	// etc
	maxTime := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	if e.to < maxTime {
		req = req.ParamUint64("to", e.to)
	}

	if e.from > 0 {
		req = req.ParamUint64("from", e.from)
	}

	if e.lessThan < math.MaxInt64 {
		req = req.ParamInt64("less_than", e.lessThan)
	}

	if e.greaterThan > math.MinInt64 {
		req = req.ParamInt64("greater_than", e.greaterThan)
	}

	if len(e.equal) > 0 {
		temp, _ := strconv.ParseInt(e.equal, 0, 64)
		req = req.ParamInt64("equals", temp)
	}

	if len(e.operator) > 0 {
		req = req.Param("operator", e.operator)

		if len(e.groupBy) > 0 {
			req = req.Param("group_by", e.groupBy)
		}
	}

	x := make(map[string]interface{})
	_, err := req.ResponseBody(&x).
		ResponseBodyHandler(func(token interface{}) error {

		if e.output == outputJson {
			output, err := json.MarshalIndent(token, "", "  ")
			fmt.Println(string(output))
			return err
		} else {
			fmt.Println(token)
			return nil
		}
	}).Execute()

	return err
}
