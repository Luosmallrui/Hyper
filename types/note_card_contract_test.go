package types

import (
	"encoding/json"
	"testing"
)

func TestNoteCardIncludesAuthorAndLikeFields(t *testing.T) {
	body, err := json.Marshal(Note{})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"nickname", "avatar", "like_count"} {
		if _, ok := payload[field]; !ok {
			t.Fatalf("note card contract missing %s: %s", field, body)
		}
	}
	if _, ok := payload["id"].(string); !ok {
		t.Fatalf("note card id must be JSON string: %s", body)
	}
}
