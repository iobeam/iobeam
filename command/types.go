package command

import (
	"encoding/json"
	"strconv"
	"fmt"
)

// Interface fo data that is posted to API, and generated from command-line input.
type Data interface {
	IsValid() bool
}

type requiredString struct {
	s *string
}

func (r *requiredString) Set(s string) error {
	r.s = &s
	return nil
}

func (r *requiredString) String() string {
	if r.IsValid() {
		return *r.s
	}
	return ""
}

func (r *requiredString) IsValid() bool {
	return r.s != nil
}

func (r *requiredString) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.s)
}

type requiredUint64 struct {
	v *uint64
}

func (r *requiredUint64) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)

	if err != nil {
		return err
	}
	r.v = &v
	
	return nil
}

func (r *requiredUint64) String() string {
	if r.IsValid() {
		return fmt.Sprintf("%d", *r.v)
	}
	return ""
}

func (r *requiredUint64) IsValid() bool {
	return r.v != nil
}

func (r *requiredUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v)
}
