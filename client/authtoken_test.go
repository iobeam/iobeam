package client

import (
	"fmt"
	"testing"
	"time"
)

func TestIsExpired(t *testing.T) {
	cases := []struct {
		// # of milliseconds to add tokens expire time
		// if negative, token will have expired
		// if positive, token will still be valid
		in   int32
		want bool
	}{
		{in: -3000, want: true},
		{in: 3000, want: false},
	}
	for _, c := range cases {
		now := time.Now()
		duration, _ := time.ParseDuration(fmt.Sprintf("%dms", c.in))
		expire := now.Add(duration).Format("2006-01-02 15:04:05 -0700")
		token := &AuthToken{
			Expires: expire,
		}
		if got, _ := token.IsExpired(); got != c.want {
			t.Errorf("IsExpired(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
