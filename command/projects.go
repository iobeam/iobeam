package command

import (
	"flag"
	"fmt"
	"iobeam.com/iobeam/client"
	"strconv"
)

type projectData struct {
	ProjectName string `json:"project_name"`
	ProjectId   uint64 `json:"project_id,omitempty"`
	// private data, not json marshalled
	isUpdate      bool
	isGet         bool
	isPermissions bool
}

func (u *projectData) IsValid() bool {
	if u.isUpdate || u.isGet {
		return len(u.ProjectName) > 0 ||
			u.ProjectId != 0
	} else if u.isPermissions {
		return u.ProjectId != 0
	}

	return len(u.ProjectName) > 0
}

func NewProjectsCommand() *Command {
	cmd := &Command{
		Name:  "project",
		Usage: "Create, get, or delete projects",
		SubCommands: Mux{
			"list":        newListProjectsCmd(),
			"get":         newGetProjectCmd(),
			"create":      newCreateProjectCmd(),
			"update":      newUpdateProjectCmd(),
			"permissions": newProjectPermissionsCmd(),
		},
	}

	return cmd
}

func newCreateOrUpdateProjectCmd(update bool, name string,
	action CommandAction) *Command {

	proj := projectData{
		isUpdate: update,
	}

	cmd := &Command{
		Name:    name,
		ApiPath: "/v1/projects",
		Usage:   name + " project",
		Data:    &proj,
		Flags:   flag.NewFlagSet(name, flag.ExitOnError),
		Action:  action,
	}

	if update {
		cmd.Flags.Uint64Var(&proj.ProjectId, "id", 0,
			"The project ID (REQUIRED)")
	}
	cmd.Flags.StringVar(&proj.ProjectName, "name", "", "The name of the new project")

	return cmd
}

func newCreateProjectCmd() *Command {
	return newCreateOrUpdateProjectCmd(false, "create", createProject)
}

func newUpdateProjectCmd() *Command {
	return newCreateOrUpdateProjectCmd(true, "update", updateProject)
}

func createProject(c *Command, ctx *Context) error {

	_, err := ctx.Client.
		Post(c.ApiPath).
		Body(c.Data).
		Expect(201).
		ResponseBody(c.Data).
		ResponseBodyHandler(func(body interface{}) error {

		project := body.(*projectData)
		fmt.Printf("The new project ID for %s is %d\n",
			project.ProjectName,
			project.ProjectId)
		return nil

	}).Execute()

	return err
}

func updateProject(c *Command, ctx *Context) error {

	u := c.Data.(*projectData)

	rsp, err := ctx.Client.
		Patch(c.ApiPath + "/" + strconv.FormatUint(u.ProjectId, 10)).
		Body(c.Data).
		Expect(200).
		Execute()

	if err == nil {
		fmt.Println("Project successfully updated")
	} else if rsp.Http().StatusCode == 204 {
		fmt.Println("Project not modified")
		return nil
	}

	return err
}

func newGetProjectCmd() *Command {

	p := projectData{
		isGet: true,
	}

	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/projects",
		Usage:   "get information about a project",
		Data:    &p,
		Flags:   flag.NewFlagSet("get", flag.ExitOnError),
		Action:  getProject,
	}

	cmd.Flags.Uint64Var(&p.ProjectId, "id", 0, "The project ID")
	cmd.Flags.StringVar(&p.ProjectName, "name", "", "Project name prefix")

	return cmd
}

func getProject(c *Command, ctx *Context) error {

	p := c.Data.(*projectData)
	var req *client.Request

	if p.ProjectId != 0 {
		req = ctx.Client.
			Get(c.ApiPath + "/" + strconv.FormatUint(p.ProjectId, 10))
	} else {
		req = ctx.Client.
			Get(c.ApiPath).
			Param("name", p.ProjectName)
	}

	type projectResult struct {
		ProjectId   uint64 `json:"project_id"`
		ProjectName string `json:"project_name"`
		Created     string
		Permissions struct {
			Admin []struct {
				UserId uint64 `json:"user_id"`
			}
			Read []struct {
				UserId uint64 `json:"user_id"`
			}
			Write []struct {
				UserId uint64 `json:"user_id"`
			}
		}
	}

	_, err := req.
		Expect(200).
		ResponseBody(new(projectResult)).
		ResponseBodyHandler(func(body interface{}) error {

		result := body.(*projectResult)
		fmt.Printf("ProjectId: %v\n"+
			"Project name: %v\n"+
			"Created: %v\n",
			result.ProjectId,
			result.ProjectName,
			result.Created)

		fmt.Printf("READ:  ")

		for _, u := range result.Permissions.Read {
			fmt.Printf("%v ", u.UserId)
		}

		fmt.Printf("\nWRITE: ")

		for _, u := range result.Permissions.Write {
			fmt.Printf("%v ", u.UserId)
		}

		fmt.Printf("\nADMIN: ")

		for _, u := range result.Permissions.Admin {
			fmt.Printf("%v ", u.UserId)
		}

		fmt.Printf("\n")

		return nil
	}).
		Execute()

	return err
}

func newListProjectsCmd() *Command {

	return &Command{
		Name:    "list",
		ApiPath: "/v1/projects",
		Usage:   "list projects",
		Action:  listProjects,
	}
}

func listProjects(c *Command, ctx *Context) error {

	type projectsResult struct {
		Projects []struct {
			ProjectId   uint64 `json:"project_id"`
			ProjectName string `json:"project_name"`
			Created     string
			Permissions struct {
				Admin bool
				Read  bool
				Write bool
			}
		}
	}

	_, err := ctx.Client.
		Get(c.ApiPath).
		Expect(200).
		ResponseBody(new(projectsResult)).
		ResponseBodyHandler(func(body interface{}) error {

		result := body.(*projectsResult)

		for _, p := range result.Projects {
			perms := ""

			if p.Permissions.Read {
				perms += "READ "
			}
			if p.Permissions.Write {
				perms += "WRITE "
			}
			if p.Permissions.Admin {
				perms += "ADMIN "
			}

			fmt.Printf("Project ID: %v\n"+
				"Project name: %v\n"+
				"Created: %v\n"+
				"Permissions: %v\n\n",
				p.ProjectId,
				p.ProjectName,
				p.Created,
				perms)
		}

		return nil
	}).Execute()

	return err
}

func newProjectPermissionsCmd() *Command {

	p := projectData{
		isPermissions: true,
	}

	cmd := &Command{
		Name:    "permissions",
		ApiPath: "/v1/projects/%v/permissions",
		Usage:   "get project permissions",
		Data:    &p,
		Flags:   flag.NewFlagSet("permissions", flag.ExitOnError),
		Action:  getProjectPermissions,
	}

	cmd.Flags.Uint64Var(&p.ProjectId, "id", 0, "The project ID")

	return cmd
}

func getProjectPermissions(c *Command, ctx *Context) error {

	p := c.Data.(*projectData)

	type permissionsResult struct {
		Permissions struct {
			Read []struct {
				UserId uint64 `json:"user_id"`
			}
			Write []struct {
				UserId uint64 `json:"user_id"`
			}
			Admin []struct {
				UserId uint64 `json:"user_id"`
			}
		}
	}

	_, err := ctx.Client.
		Get(fmt.Sprintf(c.ApiPath, strconv.FormatUint(p.ProjectId, 10))).
		Expect(200).
		ResponseBody(new(permissionsResult)).
		ResponseBodyHandler(func(body interface{}) error {

		result := body.(*permissionsResult)

		fmt.Printf("%10v %-6v", "Permission", "UserIds")

		fmt.Printf("\n%10v ", "READ")

		for _, r := range result.Permissions.Read {
			fmt.Printf("%v ", r.UserId)
		}

		fmt.Printf("\n%10v ", "WRITE")

		for _, w := range result.Permissions.Write {
			fmt.Printf("%v ", w.UserId)
		}

		fmt.Printf("\n%10v ", "ADMIN")

		for _, a := range result.Permissions.Admin {
			fmt.Printf("%v ", a.UserId)
		}

		fmt.Printf("\n")

		return nil
	}).Execute()

	return err
}
