package command

import (
	"flag"
	"fmt"
	"strings"
)

const (
	descCreateProjectId   = "Project ID this trigger belongs to (defaults to active project)."
	descCreateTriggerName = "Name of the new trigger."
	descCreateDataExpiry  = "Time (in milliseconds) after which data is considered too old to fire trigger (0 = never too old)."
	descCreateMinDelay    = "Minimum time (in milliseconds) between successive trigger firings (used to rate limit trigger events)."
)

var actionTypes = []string{"email", "http", "mqtt", "sms"}

func init() {
	flagSetNames["trigger"] = "trigger"
	baseApiPath["trigger"] = "/v1/triggers"
}

func (c *Command) newFlagSetTrigger(cmd string) *flag.FlagSet {
	return c.NewFlagSet(flagSetNames["trigger"] + " " + cmd)
}

func getUrlForTriggerId(id uint64) string {
	return getUrlForResource(baseApiPath["trigger"], id)
}

// NewTriggersCommand returns the base 'trigger' command.
func NewTriggersCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  flagSetNames["trigger"],
		Usage: "Commands for managing triggers.",
		SubCommands: Mux{
			"add-action": newAddActionTriggerCommand(ctx),
			"create":     newCreateTriggerCommand(ctx),
			"delete":     newDeleteTriggerCommand(ctx),
			"get":        newGetTriggerCommand(ctx),
			"list":       newListTriggersCommand(ctx),
			"test":       newTestTriggerCommand(ctx),
		},
	}
	cmd.NewFlagSet("iobeam " + flagSetNames["trigger"])

	return cmd
}

//
// Common data structures used for all kinds of triggers.
//

// triggerData is the main meta data for all triggers.
type triggerData struct {
	TriggerId   uint64 `json:"trigger_id,omitempty"`
	ProjectId   uint64 `json:"project_id"`
	TriggerName string `json:"trigger_name"`
	DataExpiry  uint64 `json:"data_expiry,omitempty"`
}

func (d *triggerData) IsValid() bool {
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
		i++
	}
	fmt.Println()
}

func newTrigger(name string, projectId, dataExpiry uint64, actions []triggerAction) *fullTrigger {
	ret := &fullTrigger{
		triggerData: triggerData{
			TriggerName: name,
			ProjectId:   projectId,
			DataExpiry:  dataExpiry,
		},
		Actions: make([]triggerAction, len(actions)),
	}
	copy(ret.Actions, actions)
	return ret
}

// List command data and functions

type triggerListArgs struct {
	projectId uint64
}

func (a *triggerListArgs) IsValid() bool {
	return a.projectId > 0
}

func newListTriggersCommand(ctx *Context) *Command {
	cmdStr := "list"
	a := new(triggerListArgs)
	cmd := &Command{
		Name:    cmdStr,
		ApiPath: baseApiPath["trigger"],
		Usage:   "Get all triggers for a project",
		Data:    a,
		Action:  getAllTriggers,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
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

type triggerBaseArgs struct {
	projectId   uint64
	triggerId   uint64
	triggerName string
}

func (a *triggerBaseArgs) IsValid() bool {
	return a.projectId > 0 && (a.triggerId > 0 || len(a.triggerName) > 0)
}

func (a *triggerBaseArgs) getApiPath() string {
	if a.triggerId > 0 {
		return getUrlForTriggerId(a.triggerId)
	}
	return baseApiPath["trigger"]
}

// Single get data and functions

type triggerGetArgs struct {
	triggerBaseArgs
}

func (a *triggerGetArgs) IsValid() bool {
	return a.triggerBaseArgs.IsValid()
}

func newGetTriggerCommand(ctx *Context) *Command {
	cmdStr := "get"
	a := new(triggerGetArgs)
	cmd := &Command{
		Name: cmdStr,
		// ApiPath determined by flags
		Usage:  "Get trigger matching a name or id",
		Data:   a,
		Action: getTrigger,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to get trigger from.")
	flags.Uint64Var(&a.triggerId, "id", 0, "Trigger ID to get (either this or -name must be set).")
	flags.StringVar(&a.triggerName, "name", "", "Trigger name to get (either this or -id must be set).")

	return cmd
}

func getTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*triggerGetArgs)
	t, err := _getTrigger(ctx, &args.triggerBaseArgs)
	if err == nil {
		t.Print()
	}
	return err
}

func _getTrigger(ctx *Context, args *triggerBaseArgs) (*fullTrigger, error) {
	req := ctx.Client.Get(args.getApiPath())
	if args.triggerId <= 0 {
		req.Param("name", args.triggerName)
	}

	res := new(fullTrigger)
	_, err := req.Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(res).
		ResponseBodyHandler(func(resp interface{}) error {
		return nil
	}).Execute()

	return res, err
}

// Delete data and functions

type triggerDeleteArgs struct {
	triggerBaseArgs
}

func (a *triggerDeleteArgs) IsValid() bool {
	return a.triggerBaseArgs.IsValid()
}

func newDeleteTriggerCommand(ctx *Context) *Command {
	cmdStr := "delete"
	a := new(triggerDeleteArgs)
	cmd := &Command{
		Name: cmdStr,
		// ApiPath determined by flags
		Usage:  "Delete trigger by id",
		Data:   a,
		Action: deleteTrigger,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to delete trigger from.")
	flags.Uint64Var(&a.triggerId, "id", 0, "Trigger ID to delete.")
	// TODO: Support delete by name eventually
	//flags.StringVar(&a.triggerName, "name", "", "Trigger name to get (either this or -id must be set).")

	return cmd
}

func deleteTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*triggerDeleteArgs)
	req := ctx.Client.Delete(args.getApiPath())
	if args.triggerId <= 0 {
		req.Param("name", args.triggerName)
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
	cmdStr := "test"
	a := new(triggerTestArgs)
	cmd := &Command{
		Name:    cmdStr,
		ApiPath: baseApiPath["trigger"] + "/events/test",
		Usage:   "Test that a trigger works.",
		Data:    a,
		Action:  testTrigger,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
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

// actionFunc is a function that generates a command that is based on the type
// of trigger action given.
type actionFunc func(*Context, string) *Command

func newMuxOnActionTypeCommand(ctx *Context, action, usage string, fn actionFunc) *Command {
	cmd := &Command{
		Name:        action,
		Usage:       usage,
		SubCommands: Mux{},
	}
	for _, t := range actionTypes {
		cmd.SubCommands[t] = fn(ctx, t)
	}
	cmd.newFlagSetTrigger(action)

	return cmd
}

func newCreateTriggerCommand(ctx *Context) *Command {
	return newMuxOnActionTypeCommand(ctx, "create", "Commands for adding new triggers.", newCreateTypeCommand)
}

func newAddActionTriggerCommand(ctx *Context) *Command {
	return newMuxOnActionTypeCommand(ctx, "add-action", "Commands for adding actions to triggers.", newAddActionTypeCommand)
}

// Create data and functions

type actionArgs interface {
	Valid() bool
	setFlags(flags *flag.FlagSet)
}

type createArgs struct {
	triggerData
	minDelay uint64
	data     actionArgs
}

func (a *createArgs) IsValid() bool {
	return a.triggerData.IsValid() && a.minDelay >= 0 && a.data.Valid()
}

func (a *createArgs) setCommonFlags(flags *flag.FlagSet, ctx *Context) {
	flags.Uint64Var(&a.triggerData.ProjectId, "projectId", ctx.Profile.ActiveProject, descCreateProjectId)
	flags.StringVar(&a.triggerData.TriggerName, "name", "", descCreateTriggerName)
	flags.Uint64Var(&a.triggerData.DataExpiry, "dataExpiry", 0, descCreateDataExpiry)

	flags.Uint64Var(&a.minDelay, "minDelay", 0, descCreateMinDelay)
}

func newCreateTypeCommand(ctx *Context, action string) *Command {
	var c *createArgs
	var desc string
	switch action {
	case "email":
		c = &createArgs{data: &emailActionData{To: make([]string, 1)}}
		desc = "Create a new email trigger."
	case "http":
		c = &createArgs{data: &httpActionData{}}
		desc = "Create a new HTTP trigger."
	case "mqtt":
		c = &createArgs{data: &mqttActionData{}}
		desc = "Create a new MQTT trigger."
	case "sms":
		c = &createArgs{data: &smsActionData{}}
		desc = "Create a new Twilio SMS trigger."
	default:
		panic("Unknown action type")
	}
	return newGenericTriggerCommand(ctx, c, action, desc)
}

func newGenericTriggerCommand(ctx *Context, c *createArgs, name, desc string) *Command {
	cmd := &Command{
		Name:    name,
		ApiPath: baseApiPath["trigger"],
		Usage:   desc,
		Data:    c,
		Action:  createTrigger,
	}
	flags := cmd.newFlagSetTrigger("create " + name)
	c.setCommonFlags(flags, ctx)
	c.data.setFlags(flags)

	return cmd
}

func createTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*createArgs)
	actionType := ""
	switch args.data.(type) {
	default:
		return fmt.Errorf("Unknown action type")
	case *emailActionData:
		actionType = "email"
	case *httpActionData:
		actionType = "http"
	case *mqttActionData:
		actionType = "mqtt"
	case *smsActionData:
		actionType = "sms"
	}

	actions := []triggerAction{
		{Type: actionType, MinDelay: args.minDelay, Args: args.data},
	}

	body := newTrigger(args.triggerData.TriggerName, args.triggerData.ProjectId, args.triggerData.DataExpiry, actions)
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

// Adding action data types & funcs

type addActionArgs struct {
	triggerBaseArgs
	minDelay uint64
	data     actionArgs
}

func (a *addActionArgs) IsValid() bool {
	return a.triggerBaseArgs.IsValid() && a.minDelay >= 0 && a.data.Valid()
}

func (a *addActionArgs) setCommonFlags(flags *flag.FlagSet, ctx *Context) {
	flags.Uint64Var(&a.triggerBaseArgs.projectId, "projectId", ctx.Profile.ActiveProject, descCreateProjectId)
	flags.Uint64Var(&a.triggerBaseArgs.triggerId, "id", 0, "ID of trigger to update (this or -name is REQUIRED).")
	flags.StringVar(&a.triggerBaseArgs.triggerName, "name", "", "Name of trigger to update (this or -name is REQUIRED).")

	flags.Uint64Var(&a.minDelay, "minDelay", 0, descCreateMinDelay)
}

func newAddActionTypeCommand(ctx *Context, action string) *Command {
	var c *addActionArgs
	var desc string
	switch action {
	case "email":
		c = &addActionArgs{data: &emailActionData{To: make([]string, 1)}}
		desc = "Add new email action to trigger"
	case "http":
		c = &addActionArgs{data: &httpActionData{}}
		desc = "Add new HTTP action to trigger"
	case "mqtt":
		c = &addActionArgs{data: &mqttActionData{}}
		desc = "Add new MQTT action to trigger"
	case "sms":
		c = &addActionArgs{data: &smsActionData{}}
		desc = "Add new Twilio SMS action to trigger"
	default:
		panic("Unknown action type")
	}
	return newGenericAddActionTriggerCommand(ctx, c, action, desc)
}

func newGenericAddActionTriggerCommand(ctx *Context, c *addActionArgs, name, desc string) *Command {
	cmd := &Command{
		Name: name,
		// ApiPath determined by flags
		Usage:  desc,
		Data:   c,
		Action: addAction,
	}
	flags := cmd.newFlagSetTrigger("add-action " + name)
	c.setCommonFlags(flags, ctx)
	c.data.setFlags(flags)

	return cmd
}

func addAction(c *Command, ctx *Context) error {
	args := c.Data.(*addActionArgs)
	trigger, err := _getTrigger(ctx, &args.triggerBaseArgs)
	if err != nil {
		return err
	}

	actionType := ""
	switch args.data.(type) {
	default:
		return fmt.Errorf("Unknown action type")
	case *emailActionData:
		actionType = "email"
	case *httpActionData:
		actionType = "http"
	case *mqttActionData:
		actionType = "mqtt"
	case *smsActionData:
		actionType = "sms"
	}

	trigger.Actions = append(trigger.Actions, triggerAction{Type: actionType, MinDelay: args.minDelay, Args: args.data})
	_, err = ctx.Client.
		Put(baseApiPath["trigger"]+"/"+strconv.FormatUint(trigger.TriggerId, 10)).
		Expect(200).
		ProjectToken(ctx.Profile, trigger.ProjectId).
		Body(trigger).
		Execute()
	if err == nil {
		fmt.Println("Action successfully added to trigger.")
	}
	return err
}

// ----- INDIVIDUAL ACTION TYPES BELOW ----- //

//
// HTTP data structions and functions
//

type httpActionData struct {
	URL         string `json:"url"`
	Payload     string `json:"payload"`
	AuthHeader  string `json:"auth_header"`
	ContentType string `json:"content_type"`
}

func (d *httpActionData) Valid() bool {
	return len(d.URL) > 0 && len(d.ContentType) > 0
}

func (d *httpActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.URL, "url", "", "URL to POST to when trigger is executed.")
	flags.StringVar(&d.Payload, "payload", "", "Body of POST request (optional).")
	flags.StringVar(&d.AuthHeader, "authHeader", "", "Value of 'Authorization' header of POST request, if needed (optional).")
	flags.StringVar(&d.ContentType, "contentType", "text/plain", "Content type of payload.")
}

//
// MQTT data structures and functions
//

type mqttActionData struct {
	Broker   string `json:"broker_addr"`
	Username string `json:"username"`
	Password string `json:"password"`
	QoS      int    `json:"qos"`
	Topic    string `json:"topic"`
	Payload  string `json:"payload"`
}

func (d *mqttActionData) Valid() bool {
	return len(d.Broker) > 0 && len(d.Topic) > 0 && len(d.Payload) > 0
}

func (d *mqttActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.Broker, "broker", "", "MQTT broker address to send to.")
	flags.StringVar(&d.Username, "username", "", "Username to use with MQTT broker")
	flags.StringVar(&d.Password, "password", "", "Password to use with MQTT broker")
	flags.StringVar(&d.Topic, "topic", "", "MQTT topic to post message to.")
	flags.StringVar(&d.Payload, "payload", "", "Body of the MQTT request.")
}

//
// SMS data structures and functions
//

type smsActionData struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	From       string `json:"from"`
	To         string `json:"to"`
	Payload    string `json:"message"`
}

func (d *smsActionData) Valid() bool {
	return len(d.AccountSID) > 0 && len(d.AuthToken) > 0 && len(d.From) > 0 && len(d.To) > 0 && len(d.Payload) > 0
}

func (d *smsActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.AccountSID, "accountSid", "", "Twilio account SID.")
	flags.StringVar(&d.AuthToken, "authToken", "", "Twilio authorization token.")
	flags.StringVar(&d.From, "from", "", "Phone number of the SMS sender.")
	flags.StringVar(&d.To, "to", "", "Phone number of the SMS recipient.")
	flags.StringVar(&d.Payload, "payload", "", "SMS message body.")
}

//
// Email data structures and functions
//

type emailActionData struct {
	To      []string `json:"to"`
	Subject string   `json:"subject,omitempty"`
	Payload string   `json:"payload"`
}

func (d *emailActionData) Valid() bool {
	return len(d.To) > 0 && len(d.Payload) > 0
}

func (d *emailActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.To[0], "to", "", "Email address recipient.")
	flags.StringVar(&d.Subject, "subject", "", "Email subject line.")
	flags.StringVar(&d.Payload, "payload", "", "Email message body.")
}
