package entity

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// 级联删除我自己搞,你想着哈,如果真想删真级联删除了video的表不直接废了
type Series struct {
	ID          int64          `gorm:"column:ID;primaryKey;autoIncrement"`
	ParentID    *int64         `gorm:"column:ParentID;index;comment:如果还有上层结构"` // 这里有个 OnDelete:Set CASCADE 的关联
	UserID      int64          `gorm:"column:UserID;index;comment:用户ID"`
	Layer       int64          `gorm:"column:Layer;not null;default:0"` // 这是针对series包裹时使用的,只有相同的layer才能被包在一起
	Name        string         `gorm:"column:Name;type:varchar(255);not null;index;comment:系列名称"`
	Description string         `gorm:"column:Description;type:text"`
	VideosCount int64          `gorm:"column:VideosCount;comment:计数"`
	Videos      []Video        `gorm:"foreignKey:SeriesID;references:ID;constraint:OnDelete:SET CASCADE;"`
	ViewsCount  int64          `gorm:"column:ViewsCount;not null;default:0"`
	DeletedAt   gorm.DeletedAt `gorm:"column:DeletedAt;index;comment:软删除时间"` // 这里虽说有软删除但我得写点真删除的语句
}

func (Series) TableName() string {
	return "Series"
}

type Video struct {
	// 如果有SeriesID就一定是一个剧的一集
	// 我在一次search时不想直接进行上层的再次search
	// 当进入详情页时再进行也无妨
	SeriesID    *int64  `gorm:"column:SeriesID;index;comment:父id,系列"`
	ID          int64   `gorm:"column:ID;primaryKey;autoIncrement;comment:视频唯一ID"`
	UserID      int64   `gorm:"column:UserID;index;comment:用户ID"`
	Name        string  `gorm:"column:Name;type:varchar(255);not null;index;comment:视频名称"`
	Duration    float64 `gorm:"column:Duration;comment:时长"`   // 如果是ToEpisode这里是零
	Description string  `gorm:"column:Description;type:text"` // 简单加个吧
	// Path只指到了包含 master.m3u8, stream_0.m3u8, cover.jpg, stream_0_segment_000.ts 的文件夹
	Path string `gorm:"column:Path;type:varchar(500);not null;comment:所谓的视频所在地"`

	// 这个不会矫正本地起个服务每一小时更新一次所有记录的video的viewsCount就行
	ViewsCount int64 `gorm:"column:ViewsCount;not null;default:0"`
	// Views      []View `gorm:"foreignKey:VideoID;references:ID;constraint:OnDelete:CASCADE"`

	Tags []byte `gorm:"column:Tags;type:json;comment:标签JSON数组"`
	// 简单order一下就行
	CreatedAt time.Time `gorm:"column:CreatedAt;autoCreateTime;comment:创建时间"`

	CommentsCount int64     `gorm:"column:CommentsCount;not null;default:0"`
	Comments      []Comment `gorm:"foreignKey:VideoID;references:ID;constraint:OnDelete:CASCADE"`

	// 这是被收藏的关系
	// 这里是完完全全的收藏,但收藏可能以后会根据真正db存的东西进行矫正
	CollectedVideosCount int64 `gorm:"column:CollectedVideosCount"`
	// CollectedVideos      []CollectedVideo `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:SET NULL"`
}

func (v *Video) ToDataOne() map[string]any {
	aMap := map[string]any{
		"ID":                   v.ID,
		"UserID":               v.UserID,
		"Name":                 v.Name,
		"Duration":             v.Duration,
		"Description":          v.Description,
		"Path":                 v.Path,
		"ViewsCount":           v.ViewsCount,
		"CreatedAt":            v.CreatedAt,
		"CommentsCount":        v.CommentsCount,
		"CollectedVideosCount": v.CollectedVideosCount,
	}
	aSlice := []string{}
	json.Unmarshal(v.Tags, &aSlice)
	aMap["Tags"] = aSlice
	if v.SeriesID != nil && *v.SeriesID != 0 {
		aMap["SeriesID"] = *v.SeriesID
	}
	return aMap
}

func (Video) TableName() string {
	return "Video"
}

func (v *Video) ToVideoES() *VideoES {
	tags := make([]string, 0)
	json.Unmarshal(v.Tags, &tags)
	return &VideoES{
		SeriesID:             v.SeriesID,
		ID:                   v.ID,
		UserID:               v.UserID,
		Name:                 v.Name,
		Duration:             v.Duration,
		Description:          v.Description,
		Path:                 v.Path,
		ViewsCount:           v.ViewsCount,
		Tags:                 tags,
		CreatedAt:            v.CreatedAt,
		CommentsCount:        v.CommentsCount,
		CollectedVideosCount: v.CollectedVideosCount,
	}
}

func (v *Video) ToDataMany() map[string]interface{} {
	res := map[string]interface{}{
		IDPlain:                   v.ID,
		UserIDPlain:               v.UserID,
		DurationPlain:             v.Duration,
		NamePlain:                 v.Name,
		PathPlain:                 v.Path,
		ViewsCountPlain:           v.ViewsCount,
		CommentsCountPlain:        v.CommentsCount,
		CollectedVideosCountPlain: v.CollectedVideosCount,
	}
	if v.SeriesID != nil {
		res[SeriesIDPlain] = *v.SeriesID
	}
	if len(v.Tags) != 0 {
		data := []string{}
		json.Unmarshal(v.Tags, &data)
		res[TagsPlain] = data
	}
	return res
}

type VideoES struct {
	SeriesID             *int64    `json:"SeriesID"`
	SeriesName           string    `json:"SeriesName"`
	ID                   int64     `json:"ID"`
	UserID               int64     `json:"UserID"`
	Name                 string    `json:"Name"`
	Duration             float64   `json:"Duration"`
	Description          string    `json:"Description"`
	Path                 string    `json:"Path"`
	ViewsCount           int64     `json:"ViewsCount"`
	Tags                 []string  `json:"Tags"`
	CreatedAt            time.Time `json:"CreatedAt"`
	CommentsCount        int64     `json:"CommentsCount"`
	CollectedVideosCount int64     `json:"CollectedVideosCount"`
}

func (VideoES) IndexName() string {
	return "video"
}

func (v *VideoES) ToMap() map[string]any {

	aMap := map[string]any{
		"ID":                   v.ID,
		"UserID":               v.UserID,
		"Name":                 v.Name,
		"Duration":             v.Duration,
		"Description":          v.Description,
		"Path":                 v.Path,
		"ViewsCount":           v.ViewsCount,
		"Tags":                 v.Tags,
		"CreatedAt":            v.CreatedAt,
		"CommentsCount":        v.CommentsCount,
		"CollectedVideosCount": v.CollectedVideosCount,
	}
	if v.SeriesID != nil {
		aMap["SeriesID"] = *v.SeriesID
		aMap["SeriesName"] = v.SeriesName
	}
	return aMap
}

type View struct {
	ID int64 `gorm:"column:ID;primaryKey;autoIncrement"`
	// ALTER TABLE 你的表名
	// ADD UNIQUE KEY uk_video_user (VideoID, UserID);
	// 复合唯一
	VideoID      int64     `gorm:"column:VideoID;not null;uniqueIndex:uk_video_user;comment:视频ID" validate:"required"`
	UserID       int64     `gorm:"column:UserID;not null;uniqueIndex:uk_video_user;comment:用户ID" validate:"required"`
	LastViewTime time.Time `gorm:"column:LastViewTime" validate:"required"`
	// Count int64 `gorm:"not null;default:0"`
}

func (View) TableName() string {
	return "View"
}

func (v *View) ToMap() map[string]interface{} {
	return map[string]interface{}{
		VideoIDPlain:      v.VideoID,
		UserIDPlain:       v.UserID,
		LastViewTimePlain: v.LastViewTime,
	}
}

// 这个Children也只能是一层
// 根据CreatedAt还有 VideoID,UserID,UserID2 来生成
type Comment struct {
	ID            int64     `gorm:"column:ID;primaryKey;autoIncrement"`
	ParentID      *int64    `gorm:"column:ParentID;index;comment:父视频ID（关联视频/系列）"` // 这个自关联父类
	ChildrenCount int64     `gorm:"column:ChildrenCount;comment:计数"`
	Childrens     []Comment `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
	VideoID       int64     `gorm:"column:VideoID;not null;index;comment:视频ID"`
	UserID        int64     `gorm:"column:UserID;not null;index;comment:用户ID"`
	UserID2       int64     `gorm:"column:UserID2;comment:评论对谁说"`
	Comment       string    `gorm:"column:Comment;type:varchar(2048);"`
	CreatedAt     time.Time `gorm:"column:CreatedAt;autoCreateTime;comment:创建时间"`
}

func (Comment) TableName() string {
	return "Comment"
}

func (c *Comment) ToMapMany() map[string]interface{} {
	aMap := map[string]interface{}{
		IDPlain:            c.ID,
		ChildrenCountPlain: c.ChildrenCount,
		VideoIDPlain:       c.VideoID,
		UserIDPlain:        c.UserID,
		CommentPlain:       c.Comment,
		CreatedAtPlain:     c.CreatedAt,
	}
	if c.UserID2 != 0 {
		aMap[UserID2Plain] = c.UserID2
	}
	if c.ParentID != nil {
		aMap[ParentIDPlain] = c.ParentID
	}
	return aMap
}
