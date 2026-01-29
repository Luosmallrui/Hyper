package types

import "time"

// GlobalSearchReq 搜索请求参数
type GlobalSearchReq struct {
	Keyword string `form:"keyword" binding:"required"`
	Type    int    `form:"type,default=0"` // 0-综合, 1-用户, 2-笔记, 3-活动
	Limit   int    `form:"limit,default=10"`

	// 游标字段
	UserCursor  uint64 `form:"user_cursor"`  // 对应用户 ID
	NoteCursor  uint64 `form:"note_cursor"`  // 对应笔记 ID
	PartyCursor int64  `form:"party_cursor"` // 对应活动 ID
}

// GlobalSearchResp 聚合搜索响应
type GlobalSearchResp struct {
	Users   []SearchUserItem  `json:"users,omitempty"`
	Notes   []SearchNoteItem  `json:"notes,omitempty"`
	Parties []SearchPartyItem `json:"parties,omitempty"`

	// 返回下一页需要的游标
	NextUserCursor  uint64 `json:"next_user_cursor"`
	NextNoteCursor  uint64 `json:"next_note_cursor"`
	NextPartyCursor int64  `json:"next_party_cursor"`
	HasMore         bool   `json:"has_more"`
}
type SearchUserItem struct {
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Motto    string `json:"motto"`
}

type SearchNoteItem struct {
	ID        uint64         `json:"id"`
	Title     string         `json:"title"`
	Content   string         `json:"content"` // 前端截取
	Cover     string         `json:"cover"`   // 解析出的第一张图
	Type      int            `json:"type"`
	User      SearchUserItem `json:"user"` //展示作者信息
	CreatedAt time.Time      `json:"created_at"`
}

type SearchPartyItem struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Type         string    `json:"type"`          // 商家、俱乐部、活动
	LocationName string    `json:"location_name"` // 地址名
	Price        int64     `json:"price"`
	StartTime    time.Time `json:"start_time"`
	CoverImage   string    `json:"cover_image"`
	Status       int8      `json:"status"`
}
