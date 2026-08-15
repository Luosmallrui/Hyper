package types

// NoteClassifyTopic 笔记频道分类消息 topic
const NoteClassifyTopic = "HYPER_NOTE_CLASSIFY"

// NoteImageTagTopic 图片上传后的异步 LLM 标签任务 topic。
const NoteImageTagTopic = "HYPER_NOTE_IMAGE_TAG"

// NoteClassifyMessage 笔记分类消息体
type NoteClassifyMessage struct {
	NoteID uint64 `json:"note_id"`
}

type NoteImageTagMessage struct {
	ImageID int64  `json:"image_id,string"`
	UserID  uint64 `json:"user_id"`
	URL     string `json:"url"`
}

type NoteImageTagResult struct {
	ImageID int64                      `json:"image_id,string"`
	URL     string                     `json:"url"`
	Status  string                     `json:"status"`
	Tags    []CreateOrGetTopicResponse `json:"tags"`
	Error   string                     `json:"error,omitempty"`
}
