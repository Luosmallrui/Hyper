package service

import "testing"

func TestDecodeMessageExtKeepsSnowflakeIDsAsStrings(t *testing.T) {
	ext := decodeMessageExt(`{"note_id":2085200666392793000,"note":{"id":2085200666392793000}}`)
	const want = "2085200666392793000"
	if got, _ := ext["note_id"].(string); got != want {
		t.Fatalf("note_id = %q, want %q", got, want)
	}
	note, ok := ext["note"].(map[string]interface{})
	if !ok {
		t.Fatalf("note = %#v, want object", ext["note"])
	}
	if got, _ := note["id"].(string); got != want {
		t.Fatalf("note.id = %q, want %q", got, want)
	}
}
