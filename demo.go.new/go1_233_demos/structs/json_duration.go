package structs

import (
	"encoding/json"
	"fmt"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	if d == nil {
		return fmt.Errorf("Duration: nil receiver")
	}

	s := ""
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(v)
	return nil
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
