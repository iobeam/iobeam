package command

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	keyFile = "file"
)

func init() {
	flagSetNames[keyFile] = "iobeam file"
	baseApiPath[keyFile] = "/v1/files"
}

func (c *Command) newFlagSetFile() *flag.FlagSet {
	return c.NewFlagSet(flagSetNames[keyFile] + " " + c.Name)
}

func getUrlForFileName(name string) string {
	return fmt.Sprintf("%s/%s", baseApiPath[keyFile], name)
}

// NewFilesCommand returns the base 'device' command.
func NewFilesCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  keyFile,
		Usage: "Commands for managing files on iobeam (e.g. app JARs).",
		SubCommands: Mux{
			"delete": newDeleteFileCmd(ctx),
			"list":   newListFilesCmd(ctx),
			"upload": newUploadFileCmd(ctx),
		},
	}
	cmd.NewFlagSet(flagSetNames[keyFile])

	return cmd
}

type uploadFileArgs struct {
	projectId uint64
	path      string
}

func (a *uploadFileArgs) IsValid() bool {
	return len(a.path) > 0 && a.projectId > 0
}

func newUploadFileCmd(ctx *Context) *Command {
	args := new(uploadFileArgs)

	cmd := &Command{
		Name: "upload",
		//ApiPath determined by flags
		Usage:  "Upload a file to iobeam.",
		Data:   args,
		Action: uploadFile,
	}
	flags := cmd.newFlagSetFile()
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "The ID of the project to upload the file to (defaults to active project).")
	flags.StringVar(&args.path, "path", "", "Path to file to upload.")

	return cmd
}

func getFileSha256HashString(path string) (string, error) {

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func uploadFile(c *Command, ctx *Context) error {
	args := c.Data.(*uploadFileArgs)
	_, err := _uploadFile(ctx, args)
	return err
}

// _uploadFile does the actual file uploading and returns the checksum
// of the file or an error
func _uploadFile(ctx *Context, args *uploadFileArgs) (string, error) {
	f, err := os.Open(args.path)
	if err != nil {
		fmt.Println("Could not open file for upload:")
		return "", err
	}
	defer f.Close()
	calculatedChecksum, err := getFileSha256HashString(args.path)

	if err != nil {
		fmt.Printf("Error calculating checksum:\n")
		return "", err
	}

	_, err = ctx.Client.
		Put(getUrlForFileName(filepath.Base(args.path))).
		Expect(201).
		ProjectToken(ctx.Profile, args.projectId).
		Param("checksum", calculatedChecksum).
		Param("checksum_alg", "SHA-256").
		BodyStream(f).
		Execute()

	if err == nil {
		fmt.Printf("File '%s' uploaded successfully.\n", args.path)
		return calculatedChecksum, nil
	}
	return "", err
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
		Name: "delete",
		//ApiPath determined by flags
		Usage:  "Delete a file from iobeam.",
		Data:   args,
		Action: deleteFile,
	}
	flags := cmd.newFlagSetFile()
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "The ID of the project that contains the file (defaults to active project).")
	flags.StringVar(&args.filename, "name", "", "Name of the file to delete.")

	return cmd
}

func deleteFile(c *Command, ctx *Context) error {
	args := c.Data.(*deleteFileArgs)

	_, err := ctx.Client.
		Delete(getUrlForFileName(args.filename)).
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
		ApiPath: baseApiPath[keyFile],
		Usage:   "List files for a project.",
		Data:    args,
		Action:  listFiles,
	}
	flags := cmd.newFlagSetFile()
	flags.Uint64Var(&args.projectId, "projectId", ctx.Profile.ActiveProject, "The ID of the project to get list of files from (defaults to active project).")

	return cmd
}

type checksum struct {
	Algorithm string `json:"alg"`
	Sum       string `json:"sum"`
}

func (c *checksum) Print() {
	fmt.Printf("Checksum: %s (%s)\n", c.Sum, c.Algorithm)
}

type fileInfo struct {
	Name     string   `json:"file_name"`
	Created  string   `json:"created"`
	Checksum checksum `json:"checksum"`
}

func (i *fileInfo) Print() {
	fmt.Printf("Name    : %s\n", i.Name)
	fmt.Printf("Created : %s\n", i.Created)
	i.Checksum.Print()
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
