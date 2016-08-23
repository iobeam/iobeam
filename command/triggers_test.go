package command

import (
	"strings"
	"testing"
)

const (
	testDescInvalidNoNameOrId = "invalid, no name or id"
	testDescInvalidNameLen    = "invalid, name must have len > 0"
	testDescInvalidTriggerId  = "invalid, trigger id must be > 0"
)

var casesBaseArgs = []dataTestCase{
	{
		desc: "a valid triggerBaseArgs object w/ trigger id",
		in: &triggerBaseArgs{
			projectId: 1,
			triggerId: 1,
		},
		want: true,
	},
	{
		desc: "a valid triggerBaseArgs object w/ trigger name",
		in: &triggerBaseArgs{
			projectId:   1,
			triggerName: "test",
		},
		want: true,
	},
	{
		desc: testDescInvalidNoNameOrId,
		in: &triggerBaseArgs{
			projectId: 1,
		},
		want: false,
	},
	{
		desc: testDescInvalidProjectId,
		in: &triggerBaseArgs{
			projectId:   0,
			triggerName: "test",
		},
		want: false,
	},
	{
		desc: testDescInvalidNameLen,
		in: &triggerBaseArgs{
			projectId:   1,
			triggerName: "",
		},
		want: false,
	},
	{
		desc: testDescInvalidTriggerId,
		in: &triggerBaseArgs{
			projectId: 1,
			triggerId: 0,
		},
		want: false,
	},
}

func TestTriggerBaseArgsValidity(t *testing.T) {
	runDataTestCase(t, casesBaseArgs)
}

func TestTriggerGetArgsValidity(t *testing.T) {
	cases := make([]dataTestCase, len(casesBaseArgs))
	for i, c := range casesBaseArgs {
		cases[i].desc = strings.Replace(c.desc, "triggerBaseArgs", "triggerGetArgs", -1)
		cases[i].in = &triggerGetArgs{*c.in.(*triggerBaseArgs)}
		cases[i].want = c.want
	}
	runDataTestCase(t, cases)
}

func TestTriggerDeleteArgsValidity(t *testing.T) {
	cases := make([]dataTestCase, len(casesBaseArgs))
	for i, c := range casesBaseArgs {
		cases[i].desc = strings.Replace(c.desc, "triggerBaseArgs", "triggerDeleteArgs", -1)
		cases[i].in = &triggerDeleteArgs{*c.in.(*triggerBaseArgs)}
		cases[i].want = c.want
	}
	runDataTestCase(t, cases)
}

func TestTriggerDataValidity(t *testing.T) {
	cases := []dataTestCase{
		{
			desc: testDescInvalidProjectId,
			in: &triggerData{
				TriggerId:   0,
				ProjectId:   0, // must be > 0
				TriggerName: "trigger",
				DataExpiry:  0,
			},
			want: false,
		},
		{
			desc: "invalid trigger name (none)",
			in: &triggerData{
				TriggerId:  0,
				ProjectId:  1,
				DataExpiry: 0,
			},
			want: false,
		},
		{
			desc: "valid triggerData object",
			in: &triggerData{
				TriggerId:   0,
				ProjectId:   1,
				TriggerName: "trigger",
				FireWhen:    "{{ temp }} > 20",
				DataExpiry:  0,
			},
			want: true,
		},
	}

	runDataTestCase(t, cases)
}

func TestEmailDataValidity(t *testing.T) {
	cases := []struct {
		in   *emailActionData
		want bool
	}{
		{
			in: &emailActionData{
				To:      make([]string, 0), // must have len > 0
				Payload: "test",
			},
			want: false,
		},
		{
			in: &emailActionData{
				To:      make([]string, 1),
				Payload: "",
			},
			want: false,
		},
		{
			in: &emailActionData{
				To:      make([]string, 1),
				Payload: "test",
			},
			want: true,
		},
	}

	for _, c := range cases {
		if got := c.in.Valid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHTTPDataValidity(t *testing.T) {
	cases := []struct {
		in   *httpActionData
		want bool
	}{
		{
			in: &httpActionData{
				URL:         "", // must have len > 0
				ContentType: "text/plain",
			},
			want: false,
		},
		{
			in: &httpActionData{
				URL:         "iobeam.com",
				ContentType: "",
			},
			want: false,
		},
		{
			in: &httpActionData{
				URL:         "iobeam.com",
				ContentType: "text/plain",
			},
			want: true,
		},
	}

	for _, c := range cases {
		if got := c.in.Valid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMQTTDataValidity(t *testing.T) {
	cases := []struct {
		in   *mqttActionData
		want bool
	}{
		{
			in: &mqttActionData{
				Broker:  "iobeam.com",
				Topic:   "good topic",
				Payload: "message",
			},
			want: true,
		},
		{
			in: &mqttActionData{
				Broker:  "", // must have len > 0
				Topic:   "good topic",
				Payload: "message",
			},
			want: false,
		},
		{
			in: &mqttActionData{
				Broker:  "iobeam.com",
				Topic:   "", // must have len > 0
				Payload: "message",
			},
			want: false,
		},
		{
			in: &mqttActionData{
				Broker:  "iobeam.com",
				Topic:   "good topic",
				Payload: "", // must have len > 0
			},
			want: false,
		},
	}

	for _, c := range cases {
		if got := c.in.Valid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSMSDataValidity(t *testing.T) {
	cases := []struct {
		in   *smsActionData
		want bool
	}{
		{
			in: &smsActionData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "message",
			},
			want: true,
		},
		{
			in: &smsActionData{
				AccountSID: "", // must have len > 0
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsActionData{
				AccountSID: "my id",
				AuthToken:  "", // must have len > 0
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsActionData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "", // must have len > 0
				To:         "0000000000",
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsActionData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "", // must have len > 0
				Payload:    "message",
			},
			want: false,
		},
		{
			in: &smsActionData{
				AccountSID: "my id",
				AuthToken:  "my token",
				From:       "0000000000",
				To:         "0000000000",
				Payload:    "", // must have len > 0
			},
			want: false,
		},
	}

	for _, c := range cases {
		if got := c.in.Valid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
