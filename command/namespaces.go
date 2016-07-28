package command

import (
	"fmt"
)

//{"namespace_id":57,"project_id":225,"name":"input","partitioning_field":"device_id","created":"2016-07-29T08:19:47Z","last_modified":"2016-07-29T08:19:47Z","fields":{"device_id":"STRING","double_series":"STRING"},"labels":{"device_id:distinct":true}}

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

func (ns *namespaceData) Print() {
	fmt.Printf("Name: %s\n", ns.Name)
	fmt.Printf("Id: %d\n", ns.NamespaceId)
	fmt.Printf("Partitioning field: %s\n", ns.PartitioningField)
	fmt.Printf("Created: %s\n", ns.Created)
	fmt.Printf("Last modified: %s\n", ns.LastModified)
	fmt.Printf("Fields:\n")

	for fieldName, fieldType := range ns.Fields {
		fmt.Printf("\t%s:%s\n", fieldName, fieldType)
	}

	fmt.Printf("Labels:\n")

	for labelName, labelValue := range ns.Labels {
		fmt.Printf("\t%s:%s\n", labelName, labelValue)
	}

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
		DumpRequestOut(d.dumpRequest).
		DumpResponse(d.dumpResponse).
		Expect(200).
		ResponseBody(new(namespaceResult)).
		ResponseBodyHandler(func(body interface{}) error {

			result := body.(*namespaceResult)

			for _, n := range result.Namespaces {
				n.Print()
				fmt.Print("\n")
			}

			return nil
		}).Execute()

	return err
}
