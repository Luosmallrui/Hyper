package utils

import (
	"encoding/json"
)

func unmarshalJSON[T any](src string, dst *T) error {
	if src == "" {
		return nil
	}
	return json.Unmarshal([]byte(src), dst)
}
