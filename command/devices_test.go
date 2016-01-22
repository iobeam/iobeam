package command

import "testing"

func TestBaseDeviceArgsIsValid(t *testing.T) {
	cases := []struct {
		in   *baseDeviceArgs
		want bool
	}{
		{
			in: &baseDeviceArgs{
				projectId: 1,
				id:        "did",
				name:      "dname",
			},
			want: true,
		},
		{
			in: &baseDeviceArgs{
				projectId: 1,
				id:        "did",
				name:      "",
			},
			want: true,
		},
		{
			in: &baseDeviceArgs{
				projectId: 1,
				id:        "",
				name:      "dname",
			},
			want: true,
		},
		{
			in: &baseDeviceArgs{
				projectId: 1,
				// one of these must be >= 0
				id:   "",
				name: "",
			},
			want: false,
		},
	}

	for _, c := range cases {
		if got := c.in.IsValid(); got != c.want {
			t.Errorf("IsValid(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
