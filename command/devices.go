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
	isGet    bool
	isDelete bool
	isList   bool
}

func (d *deviceData) IsValid() bool {
	if d.isUpdate {
		return d.ProjectId != 0 ||
			len(d.DeviceName) > 0 ||
			len(d.DeviceType) > 0
	} else if d.isGet || d.isDelete {
		return len(d.DeviceId) > 0
	}
	return d.ProjectId != 0
}

func NewDevicesCommand() *Command {
	cmd := &Command{
		Name:  "device",
		Usage: "Create, get, or delete devices",
		SubCommands: Mux{
			"get":    newGetDeviceCmd(),
			"create": newCreateDeviceCmd(),
			"update": newUpdateDeviceCmd(),
			"list":   newListDevicesCmd(),
			"delete": newDeleteDeviceCmd(),
		},
	}

	return cmd
}

func newCreateOrUpdateDeviceCmd(update bool, name string, action CommandAction) *Command {

	device := deviceData{
		isUpdate: update,
	}

	flags := flag.NewFlagSet("device", flag.ExitOnError)
	flags.Uint64Var(&device.ProjectId, "projectId", 0, "The project associated with the device")
	flags.StringVar(&device.DeviceName, "name", "", "The device name")
	flags.StringVar(&device.DeviceId, "id", "", "The device's identifier")
	flags.StringVar(&device.DeviceType, "type", "", "The type of device")

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

func newCreateDeviceCmd() *Command {
	return newCreateOrUpdateDeviceCmd(false, "create", createDevice)
}

func newUpdateDeviceCmd() *Command {
	return newCreateOrUpdateDeviceCmd(true, "update", updateDevice)
}

func createDevice(c *Command, ctx *Context) error {

	_, err := ctx.Client.
		Post(c.ApiPath).
		Body(c.Data).
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

	device := deviceData{
		isGet: true,
	}

	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/devices",
		Usage:   "get device information",
		Data:    &device,
		Flags:   flag.NewFlagSet("get", flag.ExitOnError),
		Action:  getDevice,
	}

	cmd.Flags.StringVar(&device.DeviceId, "id", "", "The ID of the device to query (REQUIRED)")

	return cmd
}

func getDevice(c *Command, ctx *Context) error {

	device := c.Data.(*deviceData)

	_, err := ctx.Client.Get(c.ApiPath + "/" + device.DeviceId).
		Expect(200).
		ResponseBody(device).
		ResponseBodyHandler(func(interface{}) error {

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

func newListDevicesCmd() *Command {

	device := deviceData{
		isList: true,
	}

	cmd := &Command{
		Name:    "list",
		ApiPath: "/v1/devices",
		Usage:   "list devices",
		Data:    &device,
		Flags:   flag.NewFlagSet("get", flag.ExitOnError),
		Action:  listDevices,
	}

	cmd.Flags.Uint64Var(&device.ProjectId, "id", 0, "List devices in this project (REQUIRED)")

	return cmd
}

func listDevices(c *Command, ctx *Context) error {

	type deviceList struct {
		Devices []deviceData
	}

	_, err := ctx.Client.
		Get(c.ApiPath).
		ParamUint64("project_id", c.Data.(*deviceData).ProjectId).
		Expect(200).
		ResponseBody(new(deviceList)).
		ResponseBodyHandler(func(body interface{}) error {

		list := body.(*deviceList)

		fmt.Printf("Devices in project %v\n", c.Data.(*deviceData).ProjectId)

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

	device := deviceData{
		isDelete: true,
	}

	cmd := &Command{
		Name:    "delete",
		ApiPath: "/v1/devices",
		Usage:   "delete device",
		Data:    &device,
		Flags:   flag.NewFlagSet("delete", flag.ExitOnError),
		Action:  deleteDevice,
	}

	cmd.Flags.StringVar(&device.DeviceId, "id", "", "The ID of the device to delete (REQUIRED)")

	return cmd
}

func deleteDevice(c *Command, ctx *Context) error {

	_, err := ctx.Client.
		Delete(c.ApiPath + "/" + c.Data.(*deviceData).DeviceId).
		Expect(204).
		Execute()

	if err == nil {
		fmt.Println("Device successfully deleted")
	}

	return err
}
