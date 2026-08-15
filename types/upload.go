package types

type UploadImageResp struct {
	ImageID   int64                      `json:"image_id,string"`
	Url       string                     `json:"url"`
	Width     int                        `json:"width"`
	Height    int                        `json:"height"`
	Tags      []CreateOrGetTopicResponse `json:"tags"`
	TagStatus string                     `json:"tag_status"`
}

const (
	ImageStatusUploaded int = 0 // 已上传，未绑定
	ImageStatusBound    int = 1 // 已绑定到 note
	ImageStatusDeleted  int = 2 // 已删除（逻辑删除）

	ImageTagStatusPending   int = 0 // 标签任务已入队或等待入队
	ImageTagStatusCompleted int = 1 // LLM 标签已生成
	ImageTagStatusFailed    int = 2 // 标签生成失败，不影响图片使用
)
