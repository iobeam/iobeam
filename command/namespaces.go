package command

import (
	"bytes"
	"fmt"
)

type namespaceData struct {
	ProjectId         uint64                 `json:"project_id"`
	Name              string                 `json:"name"`
	PartitioningField string                 `json:"partitioning_field"`
	Fields            map[string]string      `json:"fields"`
	NamespaceId       uint64                 `json:"namespace_id,omitempty"`
	Labels            map[string]interface{} `json:"labels,omitempty"`
	Created           string                 `json:"created,omitempty"`
	LastModified      string                 `json:"last_modified,omitempty"`
	dumpRequest       bool
	dumpResponse      bool
}

func (nsd *namespaceData) IsValid() bool {
	return nsd.ProjectId > 0
}

// NewProjectsCommand returns the base 'namespace' command.
func NewNamespaceCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "namespace",
		Usage: "Commands for managing namespaces.",
		SubCommands: Mux{
			"list": newListNamespacesCmd(ctx),
		},
	}
	cmd.NewFlagSet("iobeam namespaces")

	return cmd
}

func (nsd *namespaceData) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Name: %s\n", nsd.Name))
	buffer.WriteString(fmt.Sprintf("Id: %d\n", nsd.NamespaceId))
	buffer.WriteString(fmt.Sprintf("Partitioning field: %s\n", nsd.PartitioningField))
	buffer.WriteString(fmt.Sprintf("Created: %s\n", nsd.Created))
	buffer.WriteString(fmt.Sprintf("Last modified: %s\n", nsd.LastModified))
	buffer.WriteString(fmt.Sprintf("Fields:\n"))

	for fieldName, fieldType := range nsd.Fields {
		buffer.WriteString(fmt.Sprintf("\t%s:%s\n", fieldName, fieldType))
	}

	buffer.WriteString(fmt.Sprintf("Labels:\n"))

	for labelName, labelValue := range nsd.Labels {
		buffer.WriteString(fmt.Sprintf("\t%s:%s\n", labelName, labelValue))
	}

	return buffer.String()
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
