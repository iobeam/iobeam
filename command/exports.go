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

const maxDurationStr = "24h"

var ops = []string{opSum, opCount, opMin, opMax, opMean}
var timeFmts = []string{timeFmtSec, timeFmtMsec, timeFmtUsec, timeFmtStruct}
var outputs = []string{outputJson, outputCsv}

type setFlags map[string]struct{}

func (i *setFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *setFlags) Set(value string) error {
	if *i == nil {
		*i = map[string]struct{}{}
	}
	var empty struct{}
	(*i)[value] = empty
	return nil
}

type exportData struct {
	projectId uint64
	deviceId  string
	series    setFlags
	timeFmt   string
	output    string

	limit       uint64
	from        uint64
	to          uint64
	lessThan    int64
	greaterThan int64
	equal       string

	last string

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
	valRangeOk := e.greaterThan <= e.lessThan

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

	lastOk := len(e.last) == 0
	if !lastOk {
		temp := e.last
		if strings.Contains(e.last, "d") {
			temp = strings.Replace(e.last, "d", "h", -1)
		}
		_, err := time.ParseDuration(temp)
		lastOk = err == nil
	}

	timeOk := isInList(e.timeFmt, timeFmts)
	outputOk := isInList(e.output, outputs)

	return pidOk && limitOk && rangeOk && valRangeOk && equalOk && opOk && groupOk && lastOk && timeOk && outputOk
}

// NewExportCommand returns the base 'export' command.
func NewExportCommand(ctx *Context) *Command {
	e := new(exportData)
	pid := ctx.Profile.ActiveProject

	cmd := &Command{
		Name:    "query",
		ApiPath: "/v1/exports",
		Usage:   "Get data for projects, devices, and series.",
		Data:    e,
		Action:  getExport,
	}

	flags := cmd.NewFlagSet("iobeam query")
	maxDuration, _ := time.ParseDuration(maxDurationStr)
	maxTime := time.Now().Add(maxDuration).UnixNano() / int64(time.Millisecond)
	flags.Uint64Var(&e.projectId, "projectId", pid, "Project ID (if omitted, defaults to active project)")
	flags.StringVar(&e.deviceId, "deviceId", "", "Device ID to filter results.")
	flags.Var(&e.series, "series", "Series name(s) to filter results (flag can be used multiple times).")
	flags.Uint64Var(&e.limit, "limit", 10, "Max number of results per stream.")

	flags.Uint64Var(&e.from, "from", 0, "Min timestamp of datapoints (unix time in milliseconds)")
	flags.Uint64Var(&e.to, "to", uint64(maxTime), "Max timestamp for datapoints (unix time in milliseconds, default is now + a day)")

	flags.Int64Var(&e.lessThan, "lessThan", math.MaxInt64, "Max value for datapoints. Cannot be less than 'greaterThan' value.")
	flags.Int64Var(&e.greaterThan, "greaterThan", math.MinInt64, "Min value for datapoints. Cannot be greater than 'lessThan' value.")
	flags.StringVar(&e.equal, "equalTo", "", "Only return datapoints with this value.")

	flags.StringVar(&e.operator, "operator", "", "Aggregation function to apply to datapoints: "+strings.Join(ops, ", "))
	flags.StringVar(&e.groupBy, "groupBy", "", "Group data by [number][period], where the time period can be ms, s, m, or h. Examples of valid values: '30s', '15m', '6h'. Requires a valid operator.")

	flags.StringVar(&e.last, "last", "", "Get datapoints for the previous [number][period] timeframe. 'period' can be: ms, s, m, h, d.")

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
	if len(e.series) == 1 {
		for key := range e.series {
			reqPath += "/" + key
			break
		}
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

	if len(e.last) > 0 {
		factor := 1
		// Go does not support duration unit of 'd' for day, but we can
		// fake it by converting it to hours then multiplying by 24
		// TODO(rrk) This won't work with mixed durations like 1d5m, but this might be unlikely.
		if strings.Contains(e.last, "d") {
			factor = 24
			e.last = strings.Replace(e.last, "d", "h", -1)
		}
		duration, _ := time.ParseDuration(e.last)
		duration = time.Duration(int64(duration) * int64(factor))
		f := time.Now().Add(-1*duration).UnixNano() / int64(time.Millisecond)
		req = req.ParamInt64("from", f)
	}

	x := make(map[string]interface{})
	_, err := req.ResponseBody(&x).
		ResponseBodyHandler(func(body interface{}) error {

		if e.output == outputJson {
			if len(e.series) > 0 {
				bodyMap := *(body.(*map[string]interface{}))
				results := bodyMap["result"].([]interface{})
				keep := []interface{}{}
				for _, val := range results {
					temp := val.(map[string]interface{})
					if _, ok := e.series[temp["name"].(string)]; ok {
						keep = append(keep, val)
					}
				}
				bodyMap["result"] = keep
			}
			output, err := json.MarshalIndent(body, "", "  ")
			fmt.Println(string(output))
			return err
		} else {
			var output string
			if len(e.series) > 0 {
				for _, v := range strings.Split(body.(string), "\n") {
					if strings.Index(v, "project_id") == 0 {
						output += v
					} else if len(v) == 0 {
						continue
					} else {
						seriesName := strings.Split(v, ",")[2]
						if _, ok := e.series[seriesName]; ok {
							output += v
						} else {
							continue
						}
					}
					output += "\n"
				}
			} else {
				output = body.(string)
			}
			fmt.Println(output)
		}
		return nil
	}).Execute()

	return err
}
