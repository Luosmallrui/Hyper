package types

// NoteClassifyTopic 笔记频道分类消息 topic
const NoteClassifyTopic = "HYPER_NOTE_CLASSIFY"

// NoteClassifyMessage 笔记分类消息体
type NoteClassifyMessage struct {
	NoteID uint64 `json:"note_id"`
}
