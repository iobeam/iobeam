package command

import (

)

import (
	"fmt"
	"flag"
	"strconv"
	"cerebriq.com/cerebctl/client"
)


type projectData struct {
	ProjectName         requiredString `json:"project_name"`
	// private data, not json marshalled
	projectId           uint64
	isUpdate            bool
}

func (u *projectData) IsValid() bool {
	if u.isUpdate {
		return u.ProjectName.IsValid() &&
			u.projectId != 0
	}
	return u.ProjectName.IsValid()
}

func NewProjectsCommand() *Command {
	cmd := &Command {
		Name: "project",
		Usage: "Create, get, or delete projects",
		SubCommands: Mux {
			"list": newListProjectsCmd(),
			"get": newGetProjectCmd(),
			"create": newCreateProjectCmd(),
			"update": newUpdateProjectCmd(),			
		},
	}

	return cmd
}

func newCreateOrUpdateProjectCmd(update bool, name string,
	action CommandAction) *Command {

	proj := projectData{
		isUpdate: update,
	}
	
	cmd := &Command {
		Name: name,
		ApiPath: "/v1/projects",
		Usage: name + " project",
		Data: &proj,
		Flags: flag.NewFlagSet(name, flag.ExitOnError),	
		Action: action,
	}

	if update {
		cmd.Flags.Uint64Var(&proj.projectId, "projectId", 0,
			"The project ID (REQUIRED)")
	}
	cmd.Flags.Var(&proj.ProjectName, "name", "The name of the new project")

	return cmd
}

func newCreateProjectCmd() *Command {
	return newCreateOrUpdateProjectCmd(false, "create", createProject)
}

func newUpdateProjectCmd() *Command {
	return newCreateOrUpdateProjectCmd(true, "update", updateProject)
}

func createProject(c *Command, ctx *Context) error {

	u := c.Data.(*projectData)

	type resultData struct {
		ProjectId     uint64 `json:"project_id"`
	}

	_, err := ctx.Client.
		Post(c.ApiPath).
		Body(&u).
		Expect(201).
		ResponseBody(new(resultData)).
		ResponseBodyHandler(func(data interface{}) error {

		projectData := data.(*resultData)
		fmt.Printf("The new project ID for %s is %d\n",
			u.ProjectName.String(),
			projectData.ProjectId)
		return nil
		
	}).Execute();
	
	return err
}

func updateProject(c *Command, ctx *Context) error {

	u := c.Data.(*projectData)

	_, err := ctx.Client.
		Patch(c.ApiPath + "/" + strconv.FormatUint(u.projectId, 10)).
		Body(&u).
		Expect(200).
		Execute();

	if err == nil {
		fmt.Println("Project successfully updated")
	}
	
	return err
}


type getProjectsData struct {
	id   uint64
	name string 
}

func (u *getProjectsData) IsValid() bool {
	return u.id != 0 || len(u.name) > 0
}

func newGetProjectCmd() *Command {
	p := getProjectsData{}
	
	cmd := &Command {
		Name: "get",
		ApiPath: "/v1/projects",
		Usage: "get information about a given project",
		Data: &p,
		Flags: flag.NewFlagSet("get", flag.ExitOnError),		
		Action: getProject,
	}

	cmd.Flags.Uint64Var(&p.id, "projectId", 0, "The project ID")
	cmd.Flags.StringVar(&p.name, "name", "", "Project name prefix to search for")
	
	return cmd
}

func getProject(c *Command, ctx *Context) error {

	p := c.Data.(*getProjectsData)
	var req *client.Request
	
	if p.id != 0 {
		req = ctx.Client.
			Get(c.ApiPath + "/" + strconv.FormatUint(p.id, 10))
	} else {
		req = ctx.Client.
			Get(c.ApiPath).
			Param("name", p.name)
	}

	type projectResult struct {
		ProjectId    uint64 `json:"project_id"`
		ProjectName  string `json:"project_name"`
		Created      string
		Permissions struct {
			Admin []struct { UserId uint64 `json:"user_id"` }
			Read  []struct { UserId uint64 `json:"user_id"` }
			Write []struct { UserId uint64 `json:"user_id"` }
		}
	}
	
	_, err := req.
		Expect(200).
		ResponseBody(new(projectResult)).
		ResponseBodyHandler(func(body interface{}) error {
		
		result := body.(*projectResult)
		fmt.Printf("ProjectId: %v\n" +
			"Project name: %v\n" +
			"Created: %v\n",
			result.ProjectId,
			result.ProjectName,
			result.Created)

		fmt.Printf("READ:  ")
		for _, u := range(result.Permissions.Read) {
			fmt.Printf("%v ", u.UserId)
		}
		fmt.Printf("\nWRITE: ")

		for _, u := range(result.Permissions.Write) {
			fmt.Printf("%v ", u.UserId)
		}

		fmt.Printf("\nADMIN: ")

		for _, u := range(result.Permissions.Admin) {
			fmt.Printf("%v ", u.UserId)
		}
		fmt.Printf("\n")
		
		return nil
	}).
		Execute()
	
	return err
}

func newListProjectsCmd() *Command {

	return &Command {
		Name: "list",
		ApiPath: "/v1/projects",
		Usage: "list projects",
		Action: listProjects,
	}
}

func listProjects(c *Command, ctx *Context) error {

	type projectsResult struct {
		Projects []struct {
			ProjectId    uint64 `json:"project_id"`
			ProjectName  string `json:"project_name"`
			Created      string
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

		for _, p := range(result.Projects) {
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

			fmt.Printf("Project ID: %v\n" +
				"Project name: %v\n" +
				"Created: %v\n" +
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
