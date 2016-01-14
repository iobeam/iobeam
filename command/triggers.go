package command

import (
	"flag"
	"fmt"
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
	TriggerName string `json:"name"`
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
