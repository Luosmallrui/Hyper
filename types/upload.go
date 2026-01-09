package types

type UploadImageResp struct {
	ImageID int64  `json:"image_id"`
	Url     string `json:"url"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

const (
	ImageStatusUploaded int = 0 // 已上传，未绑定
	ImageStatusBound    int = 1 // 已绑定到 note
	ImageStatusDeleted  int = 2 // 已删除（逻辑删除）
)
