package vo

import (
	"encoding/json"
	"time"

	"github.com/guregu/null/v5"
)

type NullDate struct {
	null.Time
}

func (t NullDate) MarshalJSON() ([]byte, error) {
	if t.NullTime.Valid {
		stamp := t.NullTime.Time.Format(time.RFC3339)
		return json.Marshal(stamp)
	}

	return []byte(`null`), nil
}

type Date struct {
	time.Time
}

func (t Date) MarshalJSON() ([]byte, error) {
	stamp := t.Format(time.RFC3339)
	return json.Marshal(stamp)
}
