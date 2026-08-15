package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestImageTagResponsesSerializeSnowflakeIDsAsStrings(t *testing.T) {
	const imageID int64 = 2089000000000000000

	upload, err := json.Marshal(UploadImageResp{ImageID: imageID})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(upload), `"image_id":"2089000000000000000"`) {
		t.Fatalf("upload image_id must be a JSON string: %s", upload)
	}

	result, err := json.Marshal(NoteImageTagResult{ImageID: imageID})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result), `"image_id":"2089000000000000000"`) {
		t.Fatalf("tag result image_id must be a JSON string: %s", result)
	}
}
