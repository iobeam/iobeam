package command

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type namespaceData struct {
	ProjectId         uint64                 `json:"project_id"`
	Name              string                 `json:"namespace_name"`
	PartitioningField string                 `json:"partitioning_field"`
	Fields            map[string]string      `json:"fields"`
	NamespaceId       uint64                 `json:"namespace_id,omitempty"`
	Labels            map[string]interface{} `json:"labels"`
	Created           string                 `json:"created,omitempty"`
	LastModified      string                 `json:"last_modified,omitempty"`
	labelsStrs        setFlags
	fieldsStrs        setFlags
	dumpRequest       bool
	dumpResponse      bool
}

func (d *namespaceData) IsValid() bool {
	return d.ProjectId > 0
}

type updateNamespaceData struct {
	namespaceData
}

func (d *updateNamespaceData) IsValid() bool {
	return len(d.fieldsStrs) > 0 && d.NamespaceId > 0 && d.ProjectId > 0
}

type createNamespaceData struct {
	namespaceData
}

func (d *createNamespaceData) IsValid() bool {
	return len(d.Name) > 0 && len(d.PartitioningField) > 0 && len(d.fieldsStrs) > 0 && d.ProjectId > 0
}

func NewNamespaceCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "namespace",
		Usage: "Commands for managing namespaces.",
		SubCommands: Mux{
			"list":   newListNamespacesCmd(ctx),
			"create": newCreateNamespaceCmd(ctx),
			"update": newUpdateNamespaceCmd(ctx),
		},
	}
	cmd.NewFlagSet("iobeam namespaces")

	return cmd
}

func (d *namespaceData) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Name: %s\n", d.Name))
	buffer.WriteString(fmt.Sprintf("Id: %d\n", d.NamespaceId))
	buffer.WriteString(fmt.Sprintf("Partitioning field: %s\n", d.PartitioningField))
	buffer.WriteString(fmt.Sprintf("Created: %s\n", d.Created))
	buffer.WriteString(fmt.Sprintf("Last modified: %s\n", d.LastModified))
	buffer.WriteString(fmt.Sprintf("Fields:\n"))

	for fieldName, fieldType := range d.Fields {
		buffer.WriteString(fmt.Sprintf("\t%s:%s\n", fieldName, fieldType))
	}

	buffer.WriteString(fmt.Sprintf("Labels:\n"))

	for labelName, labelValue := range d.Labels {
		buffer.WriteString(fmt.Sprintf("\t%s:%v\n", labelName, labelValue))
	}
	return buffer.String()
}

func newUpdateNamespaceCmd(ctx *Context) *Command {

	args := new(updateNamespaceData)
	cmd := &Command{
		Name:    "update",
		ApiPath: "/v1/namespaces/",
		Usage:   "update a namespace, fields can only be added. Labels are added.",
		Data:    args,
		Action:  updateNamespace,
	}

	flags := cmd.NewFlagSet("create namespace")
	flags.Uint64Var(&args.ProjectId, "projectId", ctx.Profile.ActiveProject, "Project ID (defaults to active project).")
	flags.Uint64Var(&args.NamespaceId, "id", 0, "Namespace id.")
	flags.Var(&args.fieldsStrs, "field", "Field on form name:type (ex: temp:DOUBLE). Supported types are DOUBLE,LONG, and STRING")
	flags.Var(&args.labelsStrs, "label", "Label to set to import batch, can occur multiple times to set multiple labels (ex. device_id=\\\"myDevice\\\")")

	flags.BoolVar(&args.dumpRequest, "dumpRequest", false, "Dump the request to std out.")
	flags.BoolVar(&args.dumpResponse, "dumpResponse", false, "Dump the response to std out.")

	return cmd
}

func updateNamespace(c *Command, ctx *Context) error {
	d := c.Data.(*updateNamespaceData)

	ns := new(namespaceData)

	path := c.ApiPath + strconv.FormatUint(d.NamespaceId, 10)
	_, err := ctx.Client.
		Get(path).
		ProjectToken(ctx.Profile, d.ProjectId).
		DumpRequest(d.dumpRequest).
		DumpResponse(d.dumpResponse).
		ResponseBody(ns).
		Expect(200).
		Execute()

	if err != nil {
		return err
	}

	if len(d.labelsStrs) > 0 {
		labels, err := parseLabels(d.labelsStrs)
		if err != nil {
			return err
		}

		for label, value := range labels {
			// Overwrite labels
			ns.Labels[label] = value
		}
	}

	fields, err := parseFieldDefinitions(d.fieldsStrs)
	if err != nil {
		return err
	}

	for fieldName, fieldType := range fields {
		if _, present := ns.Fields[fieldName]; present {
			return fmt.Errorf("Update field is not supported. Field %s\n", fieldName)
		}

		ns.Fields[fieldName] = fieldType
	}

	_, err = ctx.Client.
		Put(path).
		ProjectToken(ctx.Profile, d.ProjectId).
		DumpRequest(d.dumpRequest).
		DumpResponse(d.dumpResponse).
		Body(ns).
		Expect(204).
		Execute()

	if err == nil {
		fmt.Println("Namespace updated.")
	}

	return err
}

func newCreateNamespaceCmd(ctx *Context) *Command {

	args := new(createNamespaceData)
	cmd := &Command{
		Name:    "create",
		ApiPath: "/v1/namespaces/",
		Usage:   "Create a namespace",
		Data:    args,
		Action:  createNamespace,
	}

	flags := cmd.NewFlagSet("create namespace")
	flags.Uint64Var(&args.ProjectId, "projectId", ctx.Profile.ActiveProject, "Project ID (defaults to active project).")
	flags.StringVar(&args.Name, "name", "", "Namespace name")
	flags.StringVar(&args.PartitioningField, "partitioningField", "", "Field to partition incoming data on.")
	flags.Var(&args.fieldsStrs, "field", "Field on form name:type (ex: temp:DOUBLE). Supported types are DOUBLE,LONG; and STRING")
	flags.Var(&args.labelsStrs, "label", "Label to set")

	flags.BoolVar(&args.dumpRequest, "dumpRequest", false, "Dump the request to std out.")
	flags.BoolVar(&args.dumpResponse, "dumpResponse", false, "Dump the response to std out.")

	return cmd
}

func parseLabels(flags setFlags) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for labelStr := range flags {
		labelAndValue := strings.Split(labelStr, "=")
		if len(labelAndValue) != 2 {
			return nil, fmt.Errorf("Bad label arg: %s\n", labelStr)
		}

		value, err := strToValue(labelAndValue[1])

		if err != nil {
			return nil, err
		}
		labels[labelAndValue[0]] = value
	}
	return labels, nil
}

func parseFieldDefinitions(flags setFlags) (map[string]string, error) {
	fields := make(map[string]string)
	for fieldStr := range flags {
		fieldAndType := strings.Split(fieldStr, ":")
		if len(fieldAndType) != 2 {
			return nil, fmt.Errorf("Bad field: %s\n", fieldStr)
		}

		fieldName := fieldAndType[0]
		typeStr := fieldAndType[1]
		if !IsValidTypeString(typeStr) {
			return nil, fmt.Errorf("Bad type %s on field %s", typeStr, fieldName)
		}

		fields[fieldName] = typeStr
	}
	return fields, nil
}

func createNamespace(c *Command, ctx *Context) error {

	d := c.Data.(*createNamespaceData)

	if len(d.labelsStrs) > 0 {
		labels, err := parseLabels(d.labelsStrs)
		if err != nil {
			return err
		}
		d.Labels = labels
	}

	fields, err := parseFieldDefinitions(d.fieldsStrs)
	if err != nil {
		return err
	}

	var headers *http.Header
	d.Fields = fields
	_, err = ctx.Client.
		Post(c.ApiPath).
		ProjectToken(ctx.Profile, d.ProjectId).
		DumpRequest(d.dumpRequest).
		DumpResponse(d.dumpResponse).
		Body(d).
		ResponseHeader(&headers).
		Expect(201).
		Execute()

	if err == nil {
		location := (*headers).Get("location")
		fmt.Printf("Namespace created at %s.\n", location)
	}
	return err
}

func newListNamespacesCmd(ctx *Context) *Command {

	args := new(namespaceData)
	cmd := &Command{
		Name:    "list",
		ApiPath: "/v1/namespaces/",
		Usage:   "list namespaces",
		Data:    args,
		Action:  listNamespaces,
	}

	flags := cmd.NewFlagSet("list namespaces")
	flags.Uint64Var(&args.ProjectId, "projectId", ctx.Profile.ActiveProject, "Project ID (defaults to active project).")
	flags.BoolVar(&args.dumpRequest, "dumpRequest", false, "Dump the request to std out.")
	flags.BoolVar(&args.dumpResponse, "dumpResponse", false, "Dump the response to std out.")
	return cmd
}

func listNamespaces(c *Command, ctx *Context) error {
	type namespaceResult struct {
		Namespaces []namespaceData
	}

	d := c.Data.(*namespaceData)

	_, err := ctx.Client.
		Get(c.ApiPath).
		ProjectToken(ctx.Profile, d.ProjectId).
		DumpRequest(d.dumpRequest).
		DumpResponse(d.dumpResponse).
		Expect(200).
		ResponseBody(new(namespaceResult)).
		ResponseBodyHandler(func(body interface{}) error {

			result := body.(*namespaceResult)

			for _, n := range result.Namespaces {
				fmt.Println(n.String())
				fmt.Print("\n")
			}

			return nil
		}).Execute()

	return err
}
