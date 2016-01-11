package command

import (
	"fmt"
	"sort"
	"strings"
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
		return len(d.DeviceId) > 0 &&
			(len(d.DeviceName) > 0 || len(d.DeviceType) > 0)
	}
	return d.ProjectId != 0
}

// deviceId is a simpler struct for calls that just consist of a device id
// and optionally projectId
type deviceId struct {
	id        string
	projectId uint64
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
			"delete": newDeleteDeviceCmd(ctx),
			"get":    newGetDeviceCmd(ctx),
			"list":   newListDevicesCmd(ctx),
			"update": newUpdateDeviceCmd(ctx),
		},
	}
	cmd.NewFlagSet("iobeam device")

	return cmd
}

func newCreateOrUpdateDeviceCmd(ctx *Context, update bool, name string, action CommandAction) *Command {
	device := deviceData{
		isUpdate: update,
	}

	cmd := &Command{
		Name:    name,
		ApiPath: "/v1/devices",
		Usage:   name + " device",
		Data:    &device,
		Action:  action,
	}
	flags := cmd.NewFlagSet("iobeam device " + name)
	var idDesc string
	if update {
		idDesc = "ID of the device to be updated"
	} else {
		idDesc = "Device ID, if omitted a random one will be assigned (must be > 16 chars)"
	}
	flags.StringVar(&device.DeviceId, "id", "", idDesc)
	flags.StringVar(&device.DeviceName, "name", "", "The device name")
	flags.StringVar(&device.DeviceType, "type", "", "The type of device")
	flags.Uint64Var(&device.ProjectId, "projectId", ctx.Profile.ActiveProject, "Project ID associated with the device (if omitted, defaults to active project).")

	return cmd
}

func newCreateDeviceCmd(ctx *Context) *Command {
	return newCreateOrUpdateDeviceCmd(ctx, false, "create", createDevice)
}

func newUpdateDeviceCmd(ctx *Context) *Command {
	return newCreateOrUpdateDeviceCmd(ctx, true, "update", updateDevice)
}

func createDevice(c *Command, ctx *Context) error {
	data := c.Data.(*deviceData)
	_, err := ctx.Client.
		Post(c.ApiPath).
		Expect(201).
		ProjectToken(ctx.Profile, data.ProjectId).
		Body(data).
		ResponseBody(c.Data).
		ResponseBodyHandler(func(body interface{}) error {

		device := body.(*deviceData)
		fmt.Println("New device created.")
		fmt.Printf("Device ID: %v\n", device.DeviceId)
		fmt.Printf("Device Name: %v\n", device.DeviceName)
		fmt.Println()

		return nil
	}).Execute()

	return err
}

func updateDevice(c *Command, ctx *Context) error {

	device := c.Data.(*deviceData)

	rsp, err := ctx.Client.
		Patch(c.ApiPath+"/"+device.DeviceId).
		Expect(200).
		ProjectToken(ctx.Profile, device.ProjectId).
		Body(c.Data).
		Execute()

	if err == nil {
		fmt.Println("Device successfully updated")
	} else if rsp.Http().StatusCode == 204 {
		fmt.Println("Device not modified")
		return nil
	}

	return err
}

func newGetDeviceCmd(ctx *Context) *Command {
	data := new(deviceId)

	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/devices",
		Usage:   "get device information",
		Data:    data,
		Action:  getDevice,
	}
	flags := cmd.NewFlagSet("iobeam device get")
	flags.StringVar(&data.id, "id", "", "Device ID to query (REQUIRED)")
	flags.Uint64Var(&data.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID to get devices from (if omitted, defaults to active project)")

	return cmd
}

func getDevice(c *Command, ctx *Context) error {
	data := c.Data.(*deviceId)
	path := c.ApiPath + "/" + data.id

	device := new(deviceData)
	_, err := ctx.Client.
		Get(path).
		Expect(200).
		ProjectToken(ctx.Profile, data.projectId).
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

const (
	orderName        = "name"
	orderNameReverse = "name-r"
	orderId          = "id"
	orderIdReverse   = "id-r"
	orderDate        = "date"
	orderDateReverse = "date-r"
)

var orders = []string{orderName, orderNameReverse, orderId, orderIdReverse,
	orderDate, orderDateReverse}

type listData struct {
	projectId uint64
	order     string
}

func (d *listData) IsValid() bool {
	pidOk := d.projectId > 0
	orderOk := isInList(d.order, orders)
	return pidOk && orderOk
}

func newListDevicesCmd(ctx *Context) *Command {
	data := new(listData)

	cmd := &Command{
		Name:    "list",
		ApiPath: "/v1/devices",
		Usage:   "List devices for a given project.",
		Data:    data,
		Action:  listDevices,
	}
	flags := cmd.NewFlagSet("iobeam device list")
	flags.Uint64Var(&data.projectId, "projectId", ctx.Profile.ActiveProject,
		"Project ID to get devices from (if omitted, defaults to active project)")
	flags.StringVar(&data.order, "order", orderDate,
		"Sort order for results. Valid values: date(-r), id(-r), name(-r). Values ending with -r are reverse ordering.")

	return cmd
}

type deviceSort struct {
	items []deviceData
	order string
}

func (a deviceSort) Len() int      { return len(a.items) }
func (a deviceSort) Swap(i, j int) { a.items[i], a.items[j] = a.items[j], a.items[i] }
func (a deviceSort) Less(i, j int) bool {
	switch a.order {
	case "name":
		return strings.Compare(a.items[i].DeviceName, a.items[j].DeviceName) < 0
	case "name-r":
		return strings.Compare(a.items[j].DeviceName, a.items[i].DeviceName) < 0
	case "id":
		return strings.Compare(a.items[i].DeviceId, a.items[j].DeviceId) < 0
	case "id-r":
		return strings.Compare(a.items[j].DeviceId, a.items[i].DeviceId) < 0
	case "date-r":
		return strings.Compare(a.items[j].Created, a.items[i].Created) < 0
	case "date":
		fallthrough
	default:
		return strings.Compare(a.items[i].Created, a.items[j].Created) < 0
	}
	return false
}

func listDevices(c *Command, ctx *Context) error {
	type deviceList struct {
		Devices []deviceData
	}

	cmdArgs := c.Data.(*listData)
	pid := cmdArgs.projectId

	_, err := ctx.Client.
		Get(c.ApiPath).
		ParamUint64("project_id", pid).
		Expect(200).
		ProjectToken(ctx.Profile, pid).
		ResponseBody(new(deviceList)).
		ResponseBodyHandler(func(body interface{}) error {

		list := body.(*deviceList)

		fmt.Printf("Devices in project %v\n", pid)
		fmt.Println("-----")

		sorted := &deviceSort{items: list.Devices, order: cmdArgs.order}
		sort.Sort(sorted)
		for _, device := range sorted.items {

			fmt.Printf("Name: %v\n"+
				"Device ID: %v\n"+
				"Type: %v\n"+
				"Created: %v\n\n",
				device.DeviceName,
				device.DeviceId,
				device.DeviceType,
				device.Created)
		}

		return nil
	}).Execute()

	return err
}

func newDeleteDeviceCmd(ctx *Context) *Command {
	data := new(deviceId)

	cmd := &Command{
		Name:    "delete",
		ApiPath: "/v1/devices",
		Usage:   "delete device",
		Data:    data,
		Action:  deleteDevice,
	}
	flags := cmd.NewFlagSet("iobeam device delete")
	flags.StringVar(&data.id, "id", "", "The ID of the device to delete (REQUIRED)")
	flags.Uint64Var(&data.projectId, "projectId", ctx.Profile.ActiveProject, "The ID of the project the device belongs to (defaults to active project)")

	return cmd
}

func deleteDevice(c *Command, ctx *Context) error {
	data := c.Data.(*deviceId)
	path := c.ApiPath + "/" + data.id
	_, err := ctx.Client.
		Delete(path).
		Expect(204).
		ProjectToken(ctx.Profile, data.projectId).
		Execute()

	if err == nil {
		fmt.Println("Device successfully deleted")
	}

	return err
}
