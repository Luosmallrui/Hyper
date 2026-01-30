package types

type CreateChannelResp struct {
	ChannelId int `thrift:"channel_id,1" json:"channel_id"`
}

type CreateChannelReq struct {
	Name        string `thrift:"name,1" json:"name"`
	EnName      string `thrift:"en_name,2" json:"en_name"`
	IconUrl     string `thrift:"icon_url,3" json:"icon_url"`
	Description string `thrift:"description,4" json:"description"`
	SortWeight  int32  `thrift:"sort_weight,5" json:"sort_weight"`
	ParentId    uint32 `thrift:"parent_id,6" json:"parent_id"`
}

// ListChannelsReq 获取频道列表请求
type ListChannelsReq struct {
	// ParentId 父频道ID，如果为0则查询一级频道
	// 使用 form 标签支持 GET 请求的 Query 参数
	ParentId uint32 `json:"parent_id" form:"parent_id"`

	// 如果以后需要分页，可以预留这两个字段
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// ListChannelsResp 获取频道列表响应
type ListChannelsResp struct {
	Channels []*ChannelInfo `json:"channels"`
}

// ChannelInfo 单个频道的信息详情
type ChannelInfo struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	EnName      string `json:"en_name"`
	IconUrl     string `json:"icon_url"`
	Description string `json:"description"`
	SortWeight  int32  `json:"sort_weight"`
	ParentId    uint32 `json:"parent_id"`
}
