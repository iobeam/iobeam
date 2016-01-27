package command

import (
	"fmt"
	"os"
	"path/filepath"
)

const flagSetFile = "iobeam file"

// NewFilesCommand returns the base 'device' command.
func NewFilesCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "file",
		Usage: "Commands for managing files on iobeam (e.g. app JARs).",
		SubCommands: Mux{
			"delete": newDeleteFileCmd(ctx),
			"list":   newListFilesCmd(ctx),
			"upload": newUploadFileCmd(ctx),
		},
	}
	cmd.NewFlagSet(flagSetFile)

	return cmd
}

type uploadFileArgs struct {
	projectId uint64
	path      string
	checksum  string
}

func (a *uploadFileArgs) IsValid() bool {
	return len(a.path) > 0 && len(a.checksum) > 0 && a.projectId > 0
}

func newUploadFileCmd(ctx *Context) *Command {
	args := new(uploadFileArgs)

	cmd := &Command{
		Name:    "upload",
		ApiPath: "/v1/files",
		Usage:   "Upload a file to iobeam.",
		Data:    args,
		Action:  uploadFile,
	}
	flags := cmd.NewFlagSet(flagSetFile + " upload")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "The ID of the project to upload the file to (defaults to active project).")
	flags.StringVar(&args.path, "path", "", "Path to file to upload.")
	flags.StringVar(&args.checksum, "checksum", "", "SHA-256 checksum, as a hex digest, of the file.")

	return cmd
}

func uploadFile(c *Command, ctx *Context) error {
	args := c.Data.(*uploadFileArgs)
	f, err := os.Open(args.path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = ctx.Client.
		Post(c.ApiPath+"/"+filepath.Base(args.path)).
		Expect(201).
		ProjectToken(ctx.Profile, args.projectId).
		Param("checksum", args.checksum).
		Param("checksum_alg", "SHA-256").
		BodyStream(f).
		Execute()

	if err == nil {
		fmt.Printf("File '%s' uploaded successfully.\n", args.path)
	}
	return err
}

type deleteFileArgs struct {
	projectId uint64
	filename  string
	checksum  string
}

func (a *deleteFileArgs) IsValid() bool {
	return len(a.filename) > 0 && a.projectId > 0
}

func newDeleteFileCmd(ctx *Context) *Command {
	args := new(deleteFileArgs)

	cmd := &Command{
		Name:    "delete",
		ApiPath: "/v1/files",
		Usage:   "Delete a file from iobeam.",
		Data:    args,
		Action:  deleteFile,
	}
	flags := cmd.NewFlagSet(flagSetFile + " delete")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "The ID of the project that contains the file (defaults to active project).")
	flags.StringVar(&args.filename, "name", "", "Name of the file to delete.")

	return cmd
}

func deleteFile(c *Command, ctx *Context) error {
	args := c.Data.(*deleteFileArgs)

	_, err := ctx.Client.
		Delete(c.ApiPath+"/"+args.filename).
		Expect(204).
		ProjectToken(ctx.Profile, args.projectId).
		Execute()

	if err == nil {
		fmt.Println("File successfully deleted")
	}

	return err
}

type listFilesArgs struct {
	projectId uint64
}

func (a *listFilesArgs) IsValid() bool {
	return a.projectId > 0
}

func newListFilesCmd(ctx *Context) *Command {
	args := new(listFilesArgs)

	cmd := &Command{
		Name:    "list",
		ApiPath: "/v1/files",
		Usage:   "List files for a project.",
		Data:    args,
		Action:  listFiles,
	}
	flags := cmd.NewFlagSet(flagSetFile + " list")
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "The ID of the project to get list of files from (defaults to active project).")

	return cmd
}

type checksum struct {
	Algorithm string `json:"algorithm"`
	Sum       string `json:"sum"`
}

type fileInfo struct {
	Name     string   `json:"file_name"`
	Created  string   `json:"created"`
	Checksum checksum `json:"checksum"`
}

func (i *fileInfo) Print() {
	fmt.Printf("Name    : %s\n", i.Name)
	fmt.Printf("Created : %s\n", i.Created)
	fmt.Printf("Checksum: %s (%s)\n", i.Checksum.Sum, i.Checksum.Algorithm)
}

func listFiles(c *Command, ctx *Context) error {
	type listResult struct {
		Files []fileInfo `json:"files"`
	}
	args := c.Data.(*listFilesArgs)

	_, err := ctx.Client.
		Get(c.ApiPath).
		Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(new(listResult)).
		ResponseBodyHandler(func(body interface{}) error {
		list := body.(*listResult)
		if len(list.Files) > 0 {
			for _, info := range list.Files {
				info.Print()
			}
		} else {
			fmt.Printf("No files found for project %d.\n", args.projectId)
		}

		return nil
	}).Execute()

	return err
}
