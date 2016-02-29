package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/iobeam/iobeam/client"
)

const (
	descCreateProjectId   = "Project ID this trigger belongs to (defaults to active project)."
	descCreateTriggerName = "Name of the new trigger."
	descCreateDataExpiry  = "Time (in milliseconds) after which data is considered too old to fire trigger (0 = never too old)."
	descCreateMinDelay    = "Minimum time (in milliseconds) between successive trigger firings (used to rate limit trigger events)."
)

// NewTriggersCommand returns the base 'trigger' command.
func NewTriggersCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "trigger",
		Usage: "Commands for managing triggers.",
		SubCommands: Mux{
			"create": newCreateTriggerCommand(ctx),
			"delete": newDeleteTriggerCommand(ctx),
			"get":    newGetTriggerCommand(ctx),
			"list":   newListTriggersCommand(ctx),
			"test":   newTestTriggerCommand(ctx),
		},
	}
	cmd.NewFlagSet("iobeam trigger")

	return cmd
}

//
// Common data structures used for all kinds of triggers.
//

// triggerData is the main meta data for all triggers.
type triggerData struct {
	TriggerId   uint64 `json:"trigger_id",omitempty`
	ProjectId   uint64 `json:"project_id"`
	TriggerName string `json:"trigger_name"`
	DataExpiry  uint64 `json:"data_expiry",omitempty`
}

func (d *triggerData) isTriggerMetaValid() bool {
	return d.ProjectId > 0 && len(d.TriggerName) > 0 && d.DataExpiry >= 0
}

// triggerAction is the data for a trigger action
type triggerAction struct {
	Type     string      `json:"type"`
	MinDelay uint64      `json:"min_delay"`
	Args     interface{} `json:"args"`
}

// fullTrigger is the data structure used when sending/receiving a trigger.
type fullTrigger struct {
	triggerData
	Actions []triggerAction `json:"actions"`
}

func (t *fullTrigger) Print() {
	fmt.Println("Trigger ID  :", t.TriggerId)
	fmt.Println("Trigger name:", t.TriggerName)
	fmt.Println("Project ID  :", t.ProjectId)
	fmt.Println("Data expiry :", t.DataExpiry)
	fmt.Println("Actions:")
	i := 1
	for _, a := range t.Actions {
		if i != 1 {
			fmt.Println()
		}
		fmt.Printf("  %d) Action type: %s\n", i, a.Type)
		fmt.Println("     Min delay  :", a.MinDelay)
		fmt.Printf("     Args: %v\n", a.Args)
	}
	fmt.Println()
}

// List command data and functions

type triggerListArgs struct {
	projectId uint64
}

func (a *triggerListArgs) IsValid() bool {
	return a.projectId > 0
}

func newListTriggersCommand(ctx *Context) *Command {
	a := new(triggerListArgs)
	cmd := &Command{
		Name:    "list",
		ApiPath: "/v1/triggers",
		Usage:   "Get all triggers for a project",
		Data:    a,
		Action:  getAllTriggers,
	}

	flags := cmd.NewFlagSet("list")
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to get triggers from.")

	return cmd
}

func getAllTriggers(c *Command, ctx *Context) error {
	args := c.Data.(*triggerListArgs)
	type triggersResult struct {
		Triggers []fullTrigger
	}

	_, err := ctx.Client.Get(c.ApiPath).Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(new(triggersResult)).
		ResponseBodyHandler(func(resp interface{}) error {
		results := resp.(*triggersResult)
		for _, t := range results.Triggers {
			t.Print()
		}
		return nil
	}).Execute()

	return err
}

// Single get data and functions

type triggerGetArgs struct {
	projectId   uint64
	triggerId   uint64
	triggerName string
}

func (a *triggerGetArgs) IsValid() bool {
	return a.projectId > 0 && (a.triggerId > 0 || len(a.triggerName) > 0)
}

func newGetTriggerCommand(ctx *Context) *Command {
	a := new(triggerGetArgs)
	cmd := &Command{
		Name:    "get",
		ApiPath: "/v1/triggers",
		Usage:   "Get trigger matching a name or id",
		Data:    a,
		Action:  getTrigger,
	}

	flags := cmd.NewFlagSet("get")
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to get trigger from.")
	flags.Uint64Var(&a.triggerId, "id", 0, "Trigger ID to get (either this or -name must be set).")
	flags.StringVar(&a.triggerName, "name", "", "Trigger name to get (either this or -id must be set).")

	return cmd
}

func getTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*triggerGetArgs)
	var req *client.Request
	if args.triggerId > 0 {
		req = ctx.Client.Get(c.ApiPath + "/" + strconv.FormatUint(args.triggerId, 10))
	} else {
		req = ctx.Client.Get(c.ApiPath).Param("name", args.triggerName)
	}

	_, err := req.Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(new(fullTrigger)).
		ResponseBodyHandler(func(resp interface{}) error {
		t := resp.(*fullTrigger)
		t.Print()
		return nil
	}).Execute()

	return err
}

// Delete data and functions

type triggerDeleteArgs struct {
	projectId   uint64
	triggerId   uint64
	triggerName string
}

func (a *triggerDeleteArgs) IsValid() bool {
	return a.projectId > 0 && (a.triggerId > 0 || len(a.triggerName) > 0)
}

func newDeleteTriggerCommand(ctx *Context) *Command {
	a := new(triggerDeleteArgs)
	cmd := &Command{
		Name:    "delete",
		ApiPath: "/v1/triggers",
		Usage:   "Delete trigger by id",
		Data:    a,
		Action:  deleteTrigger,
	}

	flags := cmd.NewFlagSet("delete")
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to delete trigger from.")
	flags.Uint64Var(&a.triggerId, "id", 0, "Trigger ID to delete.")
	// TODO: Support delete by name eventually
	//flags.StringVar(&a.triggerName, "name", "", "Trigger name to get (either this or -id must be set).")

	return cmd
}

func deleteTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*triggerDeleteArgs)
	var req *client.Request
	if args.triggerId > 0 {
		req = ctx.Client.Delete(c.ApiPath + "/" + strconv.FormatUint(args.triggerId, 10))
	} else {
		req = ctx.Client.Delete(c.ApiPath).Param("name", args.triggerName)
	}

	_, err := req.Expect(204).
		ProjectToken(ctx.Profile, args.projectId).
		Execute()

	if err == nil {
		fmt.Println("Device successfully deleted")
	}

	return err
}

// Test trigger and functions

type triggerTestArgs struct {
	projectId   uint64
	triggerName string
	parameters  setFlags
}

type event struct {
	EventName string                 `json:"event_name"`
	Data      map[string]interface{} `json:"data"`
}

func (a *triggerTestArgs) IsValid() bool {
	paramsOk := true
	for k := range a.parameters {
		if len(strings.SplitN(k, ",", 2)) != 2 {
			paramsOk = false
			break
		}
	}
	return a.projectId > 0 && len(a.triggerName) > 0 && paramsOk
}

func newTestTriggerCommand(ctx *Context) *Command {
	a := new(triggerTestArgs)
	cmd := &Command{
		Name:    "test",
		ApiPath: "/v1/triggers/events/test",
		Usage:   "Test that a trigger works.",
		Data:    a,
		Action:  testTrigger,
	}

	flags := cmd.NewFlagSet("test")
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID of trigger.")
	flags.StringVar(&a.triggerName, "name", "", "Trigger name to test.")
	flags.Var(&a.parameters, "param", "Parameters for trigger in form of \"param_key,param_value\" (flag can be used multiple times).")

	return cmd
}

func testTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*triggerTestArgs)
	body := event{
		EventName: args.triggerName,
		Data:      make(map[string]interface{}),
	}
	for k := range args.parameters {
		temp := strings.SplitN(k, ",", 2)
		body.Data[temp[0]] = temp[1]
	}

	_, err := ctx.Client.Put(c.ApiPath).
		Expect(204).
		ProjectToken(ctx.Profile, args.projectId).
		Body(body).
		Execute()

	return err
}

// Create data and functions

func newCreateTriggerCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  "create",
		Usage: "Commands for adding new triggers",
		SubCommands: Mux{
			"http": newHTTPTriggerCommand(ctx),
			"mqtt": newMQTTTriggerCommand(ctx),
			"sms":  newSMSTriggerCommand(ctx),
		},
	}
	cmd.NewFlagSet("config")

	return cmd
}

// newConfig generates and sends a new trigger configuration given a body
// that is pre-made by each individual handler (http, mqtt, etc).
func newConfig(body *fullTrigger, c *Command, ctx *Context) error {
	_, err := ctx.Client.Post(c.ApiPath).Expect(201).
		ProjectToken(ctx.Profile, body.ProjectId).
		Body(body).
		ResponseBody(body).
		ResponseBodyHandler(func(resp interface{}) error {
		trigger := resp.(*fullTrigger)
		fmt.Printf("Trigger '%s' created with ID: %d\n", trigger.TriggerName, trigger.TriggerId)
		return nil
	}).Execute()

	return err
}

func (d *triggerData) setCommonFlags(flags *flag.FlagSet, ctx *Context) {
	flags.Uint64Var(&d.ProjectId, "projectId", ctx.Profile.ActiveProject, descCreateProjectId)
	flags.StringVar(&d.TriggerName, "name", "", descCreateTriggerName)
	flags.Uint64Var(&d.DataExpiry, "dataExpiry", 0, descCreateDataExpiry)
}

//
// HTTP data structions and functions
//

type httpData struct {
	URL         string `json:"url"`
	Payload     string `json:"payload"`
	AuthHeader  string `json:"auth_header"`
	ContentType string `json:"content_type"`
}

func (d *httpData) isHTTPDataValid() bool {
	return len(d.URL) > 0 && len(d.ContentType) > 0
}

type httpConfigArgs struct {
	triggerData
	minDelay uint64
	data     httpData
}

func (c *httpConfigArgs) IsValid() bool {
	return c.isTriggerMetaValid() && c.minDelay >= 0 && c.data.isHTTPDataValid()
}

func newHTTPTriggerCommand(ctx *Context) *Command {
	c := new(httpConfigArgs)
	cmd := &Command{
		Name:    "http",
		ApiPath: "/v1/triggers",
		Usage:   "Create a new HTTP trigger",
		Data:    c,
		Action:  newHTTPConfig,
	}
	flags := cmd.NewFlagSet("config http")
	c.setCommonFlags(flags, ctx)
	flags.Uint64Var(&c.minDelay, "minDelay", 0, descCreateMinDelay)

	flags.StringVar(&c.data.URL, "url", "", "URL to POST to when trigger is executed.")
	flags.StringVar(&c.data.Payload, "payload", "", "Body of POST request (optional).")
	flags.StringVar(&c.data.AuthHeader, "authHeader", "", "Value of 'Authorization' header of POST request, if needed (optional).")
	flags.StringVar(&c.data.ContentType, "contentType", "text/plain", "Content type of payload.")

	return cmd
}

func newHTTPConfig(c *Command, ctx *Context) error {
	args := c.Data.(*httpConfigArgs)
	body := fullTrigger{
		Actions: []triggerAction{
			{Type: "http", MinDelay: args.minDelay, Args: args.data},
		},
	}
	body.ProjectId = args.ProjectId
	body.TriggerName = args.TriggerName
	body.DataExpiry = args.DataExpiry

	return newConfig(&body, c, ctx)
}

//
// MQTT data structures and functions
//

type mqttData struct {
	Broker   string `json:"broker_addr"`
	Username string `json:"username"`
	Password string `json:"password"`
	QoS      int    `json:"qos"`
	Topic    string `json:"topic"`
	Payload  string `json:"payload"`
}

func (d *mqttData) isMQTTDataValid() bool {
	return len(d.Broker) > 0 && len(d.Topic) > 0 && len(d.Payload) > 0
}

type mqttConfigArgs struct {
	triggerData
	minDelay uint64
	data     mqttData
}

func (c *mqttConfigArgs) IsValid() bool {
	return c.isTriggerMetaValid() && c.minDelay >= 0 && c.data.isMQTTDataValid()
}

func newMQTTTriggerCommand(ctx *Context) *Command {
	c := new(mqttConfigArgs)
	cmd := &Command{
		Name:    "mqtt",
		ApiPath: "/v1/triggers",
		Usage:   "Create a new MQTT trigger",
		Data:    c,
		Action:  newMQTTConfig,
	}

	flags := cmd.NewFlagSet("config mqtt")
	c.setCommonFlags(flags, ctx)
	flags.Uint64Var(&c.minDelay, "minDelay", 0, descCreateMinDelay)

	flags.StringVar(&c.data.Broker, "broker", "", "MQTT broker address to send to.")
	flags.StringVar(&c.data.Username, "username", "", "Username to use with MQTT broker")
	flags.StringVar(&c.data.Password, "password", "", "Password to use with MQTT broker")
	flags.StringVar(&c.data.Topic, "topic", "", "MQTT topic to post message to.")
	flags.StringVar(&c.data.Payload, "payload", "", "Body of the MQTT request.")

	return cmd
}

func newMQTTConfig(c *Command, ctx *Context) error {
	args := c.Data.(*mqttConfigArgs)
	body := fullTrigger{
		Actions: []triggerAction{
			{Type: "mqtt", MinDelay: args.minDelay, Args: args.data},
		},
	}
	body.ProjectId = args.ProjectId
	body.TriggerName = args.TriggerName
	body.DataExpiry = args.DataExpiry

	return newConfig(&body, c, ctx)
}

//
// SMS data structures and functions
//

type smsData struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	From       string `json:"from"`
	To         string `json:"to"`
	Payload    string `json:"message"`
}

func (d *smsData) isSMSDataValid() bool {
	return len(d.AccountSID) > 0 && len(d.AuthToken) > 0 && len(d.From) > 0 && len(d.To) > 0 && len(d.Payload) > 0
}

type smsConfigArgs struct {
	triggerData
	minDelay uint64
	data     smsData
}

func (c *smsConfigArgs) IsValid() bool {
	return c.isTriggerMetaValid() && c.minDelay >= 0 && c.data.isSMSDataValid()
}

func newSMSTriggerCommand(ctx *Context) *Command {
	c := new(smsConfigArgs)
	cmd := &Command{
		Name:    "sms",
		ApiPath: "/v1/triggers",
		Usage:   "Create a new Twilio SMS trigger",
		Data:    c,
		Action:  newSMSConfig,
	}

	flags := cmd.NewFlagSet("config sms")
	c.setCommonFlags(flags, ctx)
	flags.Uint64Var(&c.minDelay, "minDelay", 0, descCreateMinDelay)

	flags.StringVar(&c.data.AccountSID, "accountSid", "", "Twilio account SID.")
	flags.StringVar(&c.data.AuthToken, "authToken", "", "Twilio authorization token.")
	flags.StringVar(&c.data.From, "from", "", "Phone number of the SMS sender.")
	flags.StringVar(&c.data.To, "to", "", "Phone number of the SMS recipient.")
	flags.StringVar(&c.data.Payload, "payload", "", "SMS message body.")

	return cmd
}

func newSMSConfig(c *Command, ctx *Context) error {
	args := c.Data.(*smsConfigArgs)
	body := fullTrigger{
		Actions: []triggerAction{
			{Type: "sms", MinDelay: args.minDelay, Args: args.data},
		},
	}
	body.ProjectId = args.ProjectId
	body.TriggerName = args.TriggerName
	body.DataExpiry = args.DataExpiry

	return newConfig(&body, c, ctx)
}
