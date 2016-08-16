package command

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	opSum   = "sum"
	opCount = "count"
	opMin   = "min"
	opMax   = "max"
	opMean  = "mean"

	predicateEq = "eq"
	predicateLt = "lt"
	predicateGt = "gt"
	predicateLe = "le"
	predicateGe = "ge"

	timeFmtSec    = "sec"
	timeFmtMsec   = "msec"
	timeFmtUsec   = "usec"
	timeFmtStruct = "timeval"

	outputJson = "json"
	outputCsv  = "csv"

	maxDurationStr = "24h"
)

var predicates = []string{predicateEq, predicateLt, predicateGt, predicateLe, predicateGe}
var ops = []string{opSum, opCount, opMin, opMax, opMean}
var timeFmts = []string{timeFmtSec, timeFmtMsec, timeFmtUsec, timeFmtStruct}
var outputs = []string{outputJson, outputCsv}

type exportData struct {
	projectId uint64
	namespace string
	fields    setFlags
	time      string
	wheres    setFlags

	groupBy  string
	operator string

	limitBy      string
	limitPeriods uint64
	limit        uint64

	timeFmt string

	output string

	rawQuery     string
	dumpRequest  bool
	dumpResponse bool
}

func (e *exportData) IsValid() bool {
	pidOk := e.projectId > 0
	limitOk := e.limit > 0

	opOk := len(e.operator) == 0 || isInList(e.operator, ops)

	groupOk := len(e.groupBy) == 0
	if !groupOk {
		if len(e.fields) != 1 {
			fmt.Println("There has to be exactly one field set when doing a groupBy.")
			return false
		}
		groupOk = len(e.operator) > 0 && isInList(e.operator, ops)
	}

	timeOk := isInList(e.timeFmt, timeFmts)
	outputOk := isInList(e.output, outputs)

	return pidOk && limitOk && opOk && groupOk && timeOk && outputOk
}

// NewExportCommand returns the base 'export' command.
func NewExportCommand(ctx *Context) *Command {
	e := new(exportData)
	pid := ctx.Profile.ActiveProject

	cmd := &Command{
		Name:    "query",
		ApiPath: "/v1/data",
		Usage:   "Get data for projects, devices, and fields.",
		Data:    e,
		Action:  getExport,
	}

	flags := cmd.NewFlagSet("iobeam query")

	flags.Uint64Var(&e.projectId, "projectId", pid, "Project ID (if omitted, defaults to active project)")
	flags.StringVar(&e.namespace, "namespace", "input", "Namespace to query.")
	flags.Var(&e.fields, "field", "Name of Field(s) to project results (flag can be used multiple times).")

	flags.StringVar(&e.time, "time", "", "Expects an interval from,to where from and to can be expressions like now()-2h or an absolute UNIX timestamp that defaults to milliseconds. The to-part is optional and defaults to now()")
	flags.Var(&e.wheres, "where", "A predicate statement on a field in the format f(field, value). Multiple where statements are allowed and form the logical conjunction (AND) (supported: "+strings.Join(predicates, ", ")+").")

	flags.StringVar(&e.groupBy, "groupBy", "", "requires the operator parameter to calculate the aggregate of a field over a specific time interval and by optional field. Ex. groupBy=time(2m),myField or groupBy=time(10s) where the time period can be ms, s, m, or h. Examples of valid values: '30s', '15m', '6h'. Requires a valid operator.")
	flags.StringVar(&e.operator, "operator", "", "Aggregation function to apply to datapoints: "+strings.Join(ops, ", "))

	flags.StringVar(&e.limitBy, "limitBy", "", "Max number of results per field (ex. location,10).")
	flags.Uint64Var(&e.limitPeriods, "limitPeriods", 0, "Limits the number of time periods to return in a group_by statement")
	flags.Uint64Var(&e.limit, "limit", 10, "Max number of results.")

	flags.StringVar(&e.timeFmt, "timeFmt", "msec", "Time unit to display timestamps: "+strings.Join(timeFmts, ", "))
	flags.StringVar(&e.output, "output", "json", "Output format of the results. Valid outputs: json, csv")

	flags.BoolVar(&e.dumpRequest, "dumpRequest", false, "Dump the request to std out.")
	flags.BoolVar(&e.dumpResponse, "dumpResponse", false, "Dump the response to std out.")

	return cmd
}

// getExport fetches the requested data from the iobeam Cloud based on
// the provided projectID, namespace, and fields names.
func getExport(c *Command, ctx *Context) error {
	e := c.Data.(*exportData)

	reqPath := c.ApiPath + "/" + e.namespace + "/"
	if len(e.fields) == 1 {
		for key := range e.fields {
			reqPath += key
			break
		}
	}

	req := ctx.Client.Get(reqPath).Expect(200).
		ProjectToken(ctx.Profile, e.projectId).
		DumpRequest(e.dumpRequest).
		DumpResponse(e.dumpResponse).
		ParamUint64("limit", e.limit).
		Param("timefmt", e.timeFmt).
		Param("output", e.output)

	if len(e.time) > 0 {
		req = req.Param("time", e.time)
	}

	if len(e.operator) > 0 {
		req = req.Param("operator", e.operator)

		if len(e.groupBy) > 0 {
			req = req.Param("group_by", e.groupBy)
		}
	}

	if len(e.wheres) > 0 {
		for key := range e.wheres {
			req = req.Param("where", key)
		}
	}
	x := make(map[string]interface{})
	_, err := req.ResponseBody(&x).
		ResponseBodyHandler(func(body interface{}) error {

			if e.output == outputJson {
				if len(e.fields) > 0 && len(e.groupBy) == 0 {
					bodyMap := *(body.(*map[string]interface{}))
					results := bodyMap["result"].([]interface{})

					for _, v := range results {
						response := v.(map[string]interface{})
						fields := response["fields"].([]interface{})
						rows := response["values"].([]interface{})

						keepFields := make(map[string]bool)
						keepFieldsNames := make([]string, len(e.fields))
						keepIndexes := make(map[int]bool)

						for field, _ := range e.fields {
							keepFields[field] = true
						}

						j := 0
						for i, field := range fields {
							if keepFields[field.(string)] {
								keepIndexes[i] = true
								keepFieldsNames[j] = field.(string)
								j += 1
							} else {
								keepIndexes[i] = false
							}
						}

						filteredRows := make([]interface{}, len(rows))

						for rowIndex, row := range rows {
							filteredRow := make([]interface{}, len(keepFields))
							filteredRowIndex := 0
							for columnIndex, value := range row.([]interface{}) {
								if keepIndexes[columnIndex] {
									filteredRow[filteredRowIndex] = value
									filteredRowIndex += 1
								}
							}
							filteredRows[rowIndex] = filteredRow
						}

						response["fields"] = keepFieldsNames
						response["values"] = filteredRows
					}
				}

				output, err := json.MarshalIndent(body, "", "  ")
				fmt.Println(string(output))
				return err
			} else {
				var output string
				if len(e.fields) > 0 {

					keepFields := make(map[int]bool)

					for i, v := range strings.Split(body.(string), "\n") {
						if i == 0 {
							// This row is the column names

							for j, fieldName := range strings.Split(v, ",") {

								if j == 0 {
									// Timestamp field
									keepFields[j] = true
									output += fieldName
									continue
								}

								if _, ok := e.fields[strings.Split(fieldName, ".")[1]]; ok {
									// keep this column
									keepFields[j] = true
									output += "," + fieldName
								} else {
									keepFields[j] = false
								}
							}
							output += "\n"
							continue
						}

						for j, value := range strings.Split(v, ",") {
							if keepFields[j] == true {
								if output[len(output)-1:] != "\n" {
									output += ","
								}
								output += value

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
