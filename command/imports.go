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
	namespace string
	fields    string
	timestamp int64
	values    string
	labels    setFlags

	dumpRequest  bool
	dumpResponse bool
}

func (d *importData) IsValid() bool {
	return d.projectId > 0 && len(d.fields) > 0 && d.timestamp >= 0 && len(d.values) > 0
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
	flags.StringVar(&d.namespace, "namespace", "input", "Namespace to import to.")
	flags.StringVar(&d.fields, "fields", "", "Comma separated list of field names (REQUIRED)")
	flags.Int64Var(&d.timestamp, "time", now, "Timestamp, in milliseconds, of the data (if omitted, defaults to current time)")
	flags.StringVar(&d.values, "values", "", "Comma separated list of data values (REQUIRED)")
	flags.Var(&d.labels, "label", "Label(s) to set for import batch (ex. device_id=\\\"myDevice\\\"). Can occur multiple times to set multiple labels.")

	flags.BoolVar(&d.dumpRequest, "dumpRequest", false, "Dump the request to std out.")
	flags.BoolVar(&d.dumpResponse, "dumpResponse", false, "Dump the response to std out.")

	return cmd
}

type dataObj struct {
	Fields []string      `json:"fields"`
	Values []interface{} `json:"values"`
}

type importObj struct {
	ProjectId uint64                 `json:"project_id"`
	Data      dataObj                `json:"data"`
	Labels    map[string]interface{} `json:"labels,omitempty"`
	Namespace string                 `json:"namespace"`
}

func strToValue(value string) (interface{}, error) {

	if IsValidDouble(value) {
		doubleValue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return doubleValue, nil
		}
		return nil, err
	}

	if IsValidLong(value) {
		longValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return longValue, nil
		}
		return nil, err
	}

	if IsValidString(value) {
		// remove the quotes
		str := strings.TrimSuffix(strings.TrimPrefix(value, "\""), "\"")

		return str, nil
	}

	return nil, errors.New("Failed to determine type of series value.")
}

func strToFieldNames(s string) ([]string, bool) {

	fieldNames := strings.Split(s, ",")
	// Check if one of the fields is time
	// if time is not in the fields, we need to
	// add it.
	for _, fieldName := range fieldNames {
		if fieldName == "time" {
			return fieldNames, false
		}
	}

	return append([]string{"time"}, fieldNames...), true
}

// strToValues takes the command line arg s, the expected number of output fields numberOfFields,
// and whether the time field should be added from system time. Returns the converted values
// according to iobeam types.
func strToValues(s string, numberOfFields int, addTime bool) ([]interface{}, error) {

	fieldsToParse := numberOfFields
	if addTime {
		fieldsToParse--
	}

	values := make([]interface{}, fieldsToParse)
	for i, s := range strings.Split(s, ",") {
		if i >= fieldsToParse {
			return nil, fmt.Errorf("Too many values in %s", s)
		}

		value, err := strToValue(s)

		if err != nil {
			return nil, err
		}

		values[i] = value
	}

	if addTime {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		values = append([]interface{}{now}, values...)
	}

	return values, nil
}

func sendImport(c *Command, ctx *Context) error {
	d := c.Data.(*importData)

	fields, addTime := strToFieldNames(d.fields)

	row, err := strToValues(d.values, len(fields), addTime)

	if err != nil {
		return err
	}

	data := dataObj{
		Fields: fields,
		Values: []interface{}{row},
	}

	labels := make(map[string]interface{})

	for labelStr := range d.labels {
		// only supports string labels for now
		labelAndValue := strings.Split(labelStr, "=")
		if len(labelAndValue) != 2 {
			return fmt.Errorf("Bad label flag arg: %s\n", labelStr)
		}

		value, _ := strToValue(labelAndValue[1])
		if value == nil {
			return fmt.Errorf("Could not determine label value: %s\n", labelAndValue[1])
		}
		labels[labelAndValue[0]] = value
	}

	obj := importObj{
		ProjectId: d.projectId,
		Namespace: d.namespace,
		Data:      data,
		Labels:    labels,
	}

	_, err = ctx.Client.
		Post(c.ApiPath).
		Expect(200).
		DumpRequest(d.dumpRequest).
		DumpResponse(d.dumpResponse).
		ProjectToken(ctx.Profile, d.projectId).
		Param("fmt", "table").
		Body(obj).
		Execute()

	if err == nil {
		fmt.Println("Data successfully imported.")
	}

	return err
}
