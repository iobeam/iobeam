package command

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/iobeam/iobeam/client"
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
	if u.isGet {
		return len(u.ProjectName) > 0 || u.ProjectId != 0
	} else if u.isUpdate {
		return len(u.ProjectName) > 0 && u.ProjectId != 0
	} else if u.isPermissions {
		return u.ProjectId != 0
	}

	return len(u.ProjectName) > 0
}

// NewProjectsCommand returns the base 'project' command.
func NewProjectsCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "project",
		Usage: "Commands for managing projects.",
		SubCommands: Mux{
			"add-user":    newAddUserProjectCmd(ctx),
			"create":      newCreateProjectCmd(),
			"get":         newGetProjectCmd(ctx),
			"list":        newListProjectsCmd(),
			"permissions": newProjectPermissionsCmd(ctx),
			"switch":      newProjectSwitchCmd(),
			"token":       newGetProjectTokenCmd(ctx),
			"update":      newUpdateProjectCmd(ctx),
		},
	}
	cmd.NewFlagSet("iobeam project")

	return cmd
}

func newCreateOrUpdateProjectCmd(ctx *Context, name string, action Action) *Command {
	update := ctx != nil
	proj := projectData{
		isUpdate: update,
	}

	cmd := &Command{
		Name:    name,
		ApiPath: "/v1/projects",
		Usage:   name + " project",
		Data:    &proj,
		Action:  action,
	}

	flags := cmd.NewFlagSet("iobeam project " + name)

	if update {
		flags.Uint64Var(&proj.ProjectId, "id", ctx.Profile.ActiveProject,
			"Project ID (if omitted, defaults to active project)")
	}
	flags.StringVar(&proj.ProjectName, "name", "", "The name of the new project")

	return cmd
}

func newCreateProjectCmd() *Command {
	return newCreateOrUpdateProjectCmd(nil, "create", createProject)
}

func newUpdateProjectCmd(ctx *Context) *Command {
	return newCreateOrUpdateProjectCmd(ctx, "update", updateProject)
}

func createProject(c *Command, ctx *Context) error {

	_, err := ctx.Client.
		Post(c.ApiPath).
		Body(c.Data).
		UserToken(ctx.Profile).
		Expect(201).
		ResponseBody(c.Data).
		ResponseBodyHandler(func(body interface{}) error {

		project := body.(*projectData)
		fmt.Printf("Project '%s' created with ID: %d\n",
			project.ProjectName,
			project.ProjectId)

		fmt.Println("Acquiring project token...")
		// Get new token for project.
		tokenCmd := newGetProjectTokenCmd(ctx)
		p := tokenCmd.Data.(*projectPermissions)
		p.projectId = project.ProjectId
		p.admin = true
		p.write = true
		p.read = true
		return getProjectToken(tokenCmd, ctx)

	}).Execute()

	return err
}

func updateProject(c *Command, ctx *Context) error {

	u := c.Data.(*projectData)

	rsp, err := ctx.Client.
		Patch(c.ApiPath + "/" + strconv.FormatUint(u.ProjectId, 10)).
		Body(c.Data).
		UserToken(ctx.Profile).
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

func newGetProjectCmd(ctx *Context) *Command {

	p := projectData{
		isGet: true,
	}

	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/projects",
		Usage:   "get information about a project",
		Data:    &p,
		Action:  getProject,
	}
	flags := cmd.NewFlagSet("iobeam project get")
	flags.Uint64Var(&p.ProjectId, "id", ctx.Profile.ActiveProject, "Project ID (if omitted, defaults to active project)")
	flags.StringVar(&p.ProjectName, "name", "", "Project name")

	return cmd
}

func getProject(c *Command, ctx *Context) error {

	p := c.Data.(*projectData)
	var req *client.Request
	if len(p.ProjectName) > 0 {
		req = ctx.Client.Get(c.ApiPath).Param("name", p.ProjectName)
	} else {
		req = ctx.Client.Get(c.ApiPath + "/" + strconv.FormatUint(p.ProjectId, 10))
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
		UserToken(ctx.Profile).
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
	}).Execute()

	return err
}

func newListProjectsCmd() *Command {
	cmd := &Command{
		Name:    "list",
		ApiPath: "/v1/projects",
		Usage:   "list projects",
		Action:  listProjects,
	}
	cmd.NewFlagSet("iobeam project list")

	return cmd
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
		UserToken(ctx.Profile).
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

type addUserData struct {
	projectId uint64
	UserId    uint64 `json:"user_id"`
	Read      bool   `json:"read"`
	Write     bool   `json:"write"`
	Admin     bool   `json:"admin"`
}

func (d *addUserData) IsValid() bool {
	return d.projectId > 0 && d.UserId > 0
}

func newAddUserProjectCmd(ctx *Context) *Command {
	d := new(addUserData)
	cmd := &Command{
		Name:    "add-user",
		ApiPath: "/v1/projects/%v/permissions",
		Usage:   "Add/remove a user to a project. Setting all flags to false will remove a user from a project.",
		Data:    d,
		Action:  addUserProject,
	}
	flags := cmd.NewFlagSet("add-user")

	flags.Uint64Var(&d.projectId, "id", ctx.Profile.ActiveProject, "Project ID (if omitted, defaults to active project)")
	flags.Uint64Var(&d.UserId, "userId", 0, "User ID of user to add")
	flags.BoolVar(&d.Read, "read", false, "Give user read permission")
	flags.BoolVar(&d.Write, "write", false, "Give user write permission")
	flags.BoolVar(&d.Admin, "admin", false, "Give user admin permission")

	return cmd
}

func addUserProject(c *Command, ctx *Context) error {
	d := c.Data.(*addUserData)

	apiPath := fmt.Sprintf(c.ApiPath, strconv.FormatUint(d.projectId, 10))
	rsp, err := ctx.Client.
		Patch(apiPath).
		UserToken(ctx.Profile).
		Body(d).
		Expect(200).Execute()

	if err == nil {
		hasPerms := d.Read || d.Write || d.Admin
		if hasPerms {
			fmt.Println("User's project permissions modified.")
		} else {
			fmt.Println("User removed from project.")
		}
	} else if rsp.Http().StatusCode == 201 {
		fmt.Println("User successfully added to project.")
		return nil
	} else if rsp.Http().StatusCode == 204 {
		fmt.Println("User's project permissions unchanged.")
		return nil
	}

	return err
}

func newProjectPermissionsCmd(ctx *Context) *Command {

	p := projectData{
		isPermissions: true,
	}

	cmd := &Command{
		Name:    "permissions",
		ApiPath: "/v1/projects/%v/permissions",
		Usage:   "get project permissions",
		Data:    &p,
		Action:  getProjectPermissions,
	}

	flags := cmd.NewFlagSet("iobeam project permissions")
	flags.Uint64Var(&p.ProjectId, "id", ctx.Profile.ActiveProject, "Project ID (if omitted, defaults to active project)")

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
		UserToken(ctx.Profile).
		Expect(200).
		ResponseBody(new(permissionsResult)).
		ResponseBodyHandler(func(body interface{}) error {

		result := body.(*permissionsResult)

		fmt.Printf("%10v | %-6v", "Permission", "UserIds")

		fmt.Printf("\n%10v | ", "READ")

		for _, r := range result.Permissions.Read {
			fmt.Printf("%v ", r.UserId)
		}

		fmt.Printf("\n%10v | ", "WRITE")

		for _, w := range result.Permissions.Write {
			fmt.Printf("%v ", w.UserId)
		}

		fmt.Printf("\n%10v | ", "ADMIN")

		for _, a := range result.Permissions.Admin {
			fmt.Printf("%v ", a.UserId)
		}

		fmt.Printf("\n")

		return nil
	}).Execute()

	return err
}

type projectId struct {
	pid uint64
}

func (p *projectId) IsValid() bool {
	return p.pid > 0
}

func newProjectSwitchCmd() *Command {
	p := projectId{}

	cmd := &Command{
		Name:    "switch",
		ApiPath: "/v1/projects/%v/permissions",
		Usage:   "Switch to a different project",
		Data:    &p,
		Action:  switchProject,
	}

	flags := cmd.NewFlagSet("iobeam project switch")
	flags.Uint64Var(&p.pid, "id", 0, "Project ID to switch to (required)")

	return cmd
}

func readBoolean(prompt string, reader *bufio.Reader) bool {
	ret := false
	for true {
		fmt.Printf(prompt)
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			temp := strings.ToLower(strings.TrimSpace(string(line)))
			if temp == "t" || temp == "true" || temp == "y" || temp == "yes" {
				ret = true
			}
			break
		}
	}
	return ret
}

func switchProject(c *Command, ctx *Context) error {
	p := c.Data.(*projectId)
	if p.pid == ctx.Profile.ActiveProject {
		fmt.Println("Already using project", p.pid)
		return nil
	}

	token, err := client.ReadProjToken(ctx.Profile, p.pid)
	if err != nil {
		fmt.Println("No token locally, need to fetch one...")
		tokenCmd := newGetProjectTokenCmd(ctx)
		pdata := tokenCmd.Data.(*projectPermissions)
		pdata.projectId = p.pid

		bio := bufio.NewReader(os.Stdin)
		pdata.read = readBoolean("Read permission? (t)rue/(f)alse: ", bio)
		pdata.write = readBoolean("Write permission? (t)rue/(f)alse: ", bio)
		pdata.admin = readBoolean("Admin permission? (t)rue/(f)alse: ", bio)
		return getProjectToken(tokenCmd, ctx)
	} else {
		expired, _ := token.IsExpired()
		if expired {
			token, err = token.Refresh(ctx.Client, ctx.Profile)
			if err != nil {
				fmt.Println("WARNING: Token is expired and could not be refreshed:")
				return err
			}
		}
		err = ctx.Profile.UpdateActiveProject(token.ProjectId)
		if err != nil {
			fmt.Printf("Could not update active project: %s\n", err)
			return err
		}
		fmt.Printf("Switched to project %v\n", p.pid)
		fmt.Printf("-----")
		printProjectToken(token)
	}

	return err
}
