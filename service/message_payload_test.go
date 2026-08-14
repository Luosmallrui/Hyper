package service

import (
	"Hyper/types"
	"strings"
	"testing"
)

func TestNormalizeMessagePayloadImage(t *testing.T) {
	message := &types.Message{
		MsgType: types.MsgTypeImage,
		Ext: map[string]interface{}{
			types.ExtKeyImageURL: " https://cdn.hypercn.cn/chat/2026/08/14/image.png ",
		},
	}

	if err := normalizeMessagePayload(message); err != nil {
		t.Fatalf("normalize image message: %v", err)
	}
	const wantURL = "https://cdn.hypercn.cn/chat/2026/08/14/image.png"
	if message.Content != wantURL {
		t.Fatalf("content = %q, want %q", message.Content, wantURL)
	}
	if got, _ := message.Ext[types.ExtKeyImageURL].(string); got != wantURL {
		t.Fatalf("ext.image_url = %q, want %q", got, wantURL)
	}
}

func TestNormalizeMessagePayloadRejectsInvalidImageURL(t *testing.T) {
	err := normalizeMessagePayload(&types.Message{
		MsgType: types.MsgTypeImage,
		Content: "file:///tmp/image.png",
	})
	if err == nil || !strings.Contains(err.Error(), "image_url") {
		t.Fatalf("error = %v, want image_url validation error", err)
	}
}

func TestNormalizeMessagePayloadRequiresTextContent(t *testing.T) {
	err := normalizeMessagePayload(&types.Message{MsgType: types.MsgTypeText, Content: "  "})
	if err == nil || !strings.Contains(err.Error(), "不能为空") {
		t.Fatalf("error = %v, want empty content error", err)
	}
}
