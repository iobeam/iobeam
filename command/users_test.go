package command

import "testing"

func TestCreateUserValidity(t *testing.T) {
	cases := []dataTestCase{
		{
			desc: "a valid userData object for creating",
			in: &userData{
				Email:  "test@iobeam.com",
				Invite: "invite-code",
			},
			want: true,
		},
		{
			desc: "invalid, missing email",
			in: &userData{
				Invite: "invite-code",
			},
			want: false,
		},
		{
			desc: "invalid, missing invite",
			in: &userData{
				Email: "test@iobeam.com",
			},
			want: false,
		},
	}

	runDataTestCase(t, cases)
}
