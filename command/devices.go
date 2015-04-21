package command

import (
	"flag"
	"fmt"
)

type deviceData struct {
	ProjectId  uint64 `json:"project_id"`
	DeviceId   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
	Created    string `json:"created,omitempty"`
	// Private fields, not marshalled into JSON
	isUpdate bool
}

func (d *deviceData) IsValid() bool {
	if d.isUpdate {
		return d.ProjectId != 0 &&
			(len(d.DeviceName) > 0 || len(d.DeviceType) > 0)
	}
	return d.ProjectId != 0
}

// deviceId is a simpler struct for calls that just consist of a device id.
type deviceId struct {
	id string
}

func (d *deviceId) IsValid() bool {
	return len(d.id) > 0
}

// NewDevicesCommand returns the base 'device' command.
func NewDevicesCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "device",
		Usage: "Commands for managing devices.",
		SubCommands: Mux{
			"create": newCreateDeviceCmd(ctx),
			"delete": newDeleteDeviceCmd(),
			"get":    newGetDeviceCmd(),
			"list":   newListDevicesCmd(ctx),
			"update": newUpdateDeviceCmd(ctx),
		},
	}

	return cmd
}

func newCreateOrUpdateDeviceCmd(ctx *Context, update bool, name string, action CommandAction) *Command {

	device := deviceData{
		isUpdate: update,
	}

	flags := flag.NewFlagSet("device", flag.ExitOnError)
	flags.Uint64Var(&device.ProjectId, "projectId", ctx.Profile.ActiveProject, "Project ID associated with the device (if omitted, defaults to active project).")
	flags.StringVar(&device.DeviceName, "name", "", "The device name")
	flags.StringVar(&device.DeviceType, "type", "", "The type of device")
	if !update {
		flags.StringVar(&device.DeviceId, "id", "", "The device's identifier")
	}

	cmd := &Command{
		Name:    name,
		ApiPath: "/v1/devices",
		Usage:   name + " device",
		Data:    &device,
		Flags:   flags,
		Action:  action,
	}

	return cmd
}

func newCreateDeviceCmd(ctx *Context) *Command {
	return newCreateOrUpdateDeviceCmd(ctx, false, "create", createDevice)
}

func newUpdateDeviceCmd(ctx *Context) *Command {
	return newCreateOrUpdateDeviceCmd(ctx, true, "update", updateDevice)
}

func createDevice(c *Command, ctx *Context) error {

	_, err := ctx.Client.
		Post(c.ApiPath).
		Body(c.Data).
		UserToken(ctx.Profile).
		Expect(201).
		ResponseBody(c.Data).
		ResponseBodyHandler(func(body interface{}) error {

		device := body.(*deviceData)
		fmt.Printf("The new device ID is %v\n",
			device.DeviceId)

		return nil
	}).Execute()

	return err
}

func updateDevice(c *Command, ctx *Context) error {

	device := c.Data.(*deviceData)

	rsp, err := ctx.Client.
		Patch(c.ApiPath + "/" + device.DeviceId).
		Body(c.Data).
		UserToken(ctx.Profile).
		Expect(200).
		Execute()

	if err == nil {
		fmt.Println("Device successfully updated")
	} else if rsp.Http().StatusCode == 204 {
		fmt.Println("Device not modified")
		return nil
	}

	return err
}

func newGetDeviceCmd() *Command {
	data := new(deviceId)

	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/devices",
		Usage:   "get device information",
		Data:    data,
		Flags:   flag.NewFlagSet("get", flag.ExitOnError),
		Action:  getDevice,
	}

	cmd.Flags.StringVar(&data.id, "id", "", "Device ID to query (REQUIRED)")

	return cmd
}

func getDevice(c *Command, ctx *Context) error {
	id := c.Data.(*deviceId).id

	device := new(deviceData)
	_, err := ctx.Client.Get(c.ApiPath + "/" + id).
		UserToken(ctx.Profile).
		Expect(200).
		ResponseBody(device).
		ResponseBodyHandler(func(body interface{}) error {
		device = body.(*deviceData)
		fmt.Printf("Device name: %v\n"+
			"Device ID: %v\n"+
			"Project ID: %v\n"+
			"Type: %v\n"+
			"Created: %v\n",
			device.DeviceName,
			device.DeviceId,
			device.ProjectId,
			device.DeviceType,
			device.Created)

		return nil
	}).Execute()

	return err
}

type listData struct {
	projectId uint64
}

func (d *listData) IsValid() bool {
	return d.projectId != 0
}

func newListDevicesCmd(ctx *Context) *Command {
	data := new(listData)

	cmd := &Command{
		Name:    "list",
		ApiPath: "/v1/devices",
		Usage:   "List devices for a given project.",
		Data:    data,
		Flags:   flag.NewFlagSet("list", flag.ExitOnError),
		Action:  listDevices,
	}

	cmd.Flags.Uint64Var(&data.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID to get devices from (if omitted, defaults to active project)")

	return cmd
}

func listDevices(c *Command, ctx *Context) error {
	type deviceList struct {
		Devices []deviceData
	}
	pid := c.Data.(*listData).projectId

	_, err := ctx.Client.
		Get(c.ApiPath).
		ParamUint64("project_id", pid).
		UserToken(ctx.Profile).
		Expect(200).
		ResponseBody(new(deviceList)).
		ResponseBodyHandler(func(body interface{}) error {

		list := body.(*deviceList)

		fmt.Printf("Devices in project %v\n", pid)

		for _, device := range list.Devices {

			fmt.Printf("\nName: %v\n"+
				"Device ID: %v\n"+
				"Type: %v\n"+
				"Created: %v\n",
				device.DeviceName,
				device.DeviceId,
				device.DeviceType,
				device.Created)
		}

		return nil
	}).Execute()

	return err
}

func newDeleteDeviceCmd() *Command {
	data := new(deviceId)

	cmd := &Command{
		Name:    "delete",
		ApiPath: "/v1/devices",
		Usage:   "delete device",
		Data:    data,
		Flags:   flag.NewFlagSet("delete", flag.ExitOnError),
		Action:  deleteDevice,
	}

	cmd.Flags.StringVar(&data.id, "id", "", "The ID of the device to delete (REQUIRED)")

	return cmd
}

func deleteDevice(c *Command, ctx *Context) error {

	_, err := ctx.Client.
		Delete(c.ApiPath + "/" + c.Data.(*deviceId).id).
		UserToken(ctx.Profile).
		Expect(204).
		Execute()

	if err == nil {
		fmt.Println("Device successfully deleted")
	}

	return err
}
