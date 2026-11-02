package entity

import (
	"mime/multipart"
	"strconv"
)

const emptyFile = "emptyFile"

// 四种查询
// UserID
// Name
// TagsJson
// 这里可以直接冲seriesID获得所有
// 其他就是分页跟限制或排序
type VideoSearchMany struct {
	Key       string   `form:"Key" validate:"required"`
	Type      string   `form:"Type" validate:"oneof=SeriesID UserID Name Tags"`
	Tags      []string `form:"-"`
	LimitTime int64    `form:"LimitTime"`
	Page      int      `form:"Page" validate:"gte=0"`
	OrderBy   string   `form:"OrderBy" validate:"omitempty,oneof=CreatedAt ViewsCount CollectedCount CommentsCount"`
	DESC      bool     `form:"DESC"`
}

func (v *VideoSearchMany) Fun() (err error) {
	switch v.Type {
	case TagsPlain:
		v.Tags, err = ToSliceString(v.Key)
	case SeriesIDPlain, UserIDPlain:
		_, err = strconv.ParseInt(v.Key, 10, 64)
	}
	return
}

// 这里可以创建video也可以是episode
// 必带文件,直接给video数据库加东西
type VideoFileUpload struct {
	File   *multipart.FileHeader `form:"File" validate:"required"` // // 这里可不有文件如果type是ToEpisode
	SHA256 string                `form:"SHA256" validate:"required,len=64"`
	// 这里有video文件类型的问题
	// Format string `form:"Format" validate:"required"`
	// 可替换video
	// 基础字段
	// 如果是系列的一个视频
	Name     string `form:"Name" validate:"required"`
	SeriesID int64  `form:"SeriesID"`

	Description string   `form:"Description" validate:"required"`
	TagsJson    string   `form:"Tags" validate:"required"`
	Tags        []string `form:"-"`
}

func (v *VideoFileUpload) Fun() (err error) {
	if v.TagsJson != "" {
		v.Tags, err = ToSliceString(v.TagsJson)
	}
	return
}

// 可以改Description
// 可以删除
// 可改name
// tags
type VideoUpdate struct {
	ID          int64    `form:"ID" validate:"required"`
	Tags        []string `form:"-"`
	TagsJson    string   `form:"Tags" validate:"required_without_all=Name Description Delete"`
	Name        string   `form:"Name" validate:"required_without_all=TagsJson Description Delete"`
	Description string   `form:"Description" validate:"required_without_all=TagsJson Name Delete"`
	Delete      bool     `form:"Delete" validate:"required_without_all=TagsJson Name Description"`
}

func (v *VideoUpdate) Fun() (err error) {
	if v.TagsJson != "" {
		v.Tags, err = ToSliceString(v.TagsJson)
	}
	return
}

// 这个可以对其内部进行parentID指定或者指定video里的seriesID
type VideoSeriesCreate struct {
	BaseOn      string  `form:"BaseOn" validate:"oneof=Series Video"`
	Name        string  `form:"Name" validate:"required"`
	Description string  `form:"Description" validate:"required"`
	IDsJson     string  `form:"IDs" validate:"required"`
	IDs         []int64 `form:"-"` // 这里能指定关联的id也就是作为一个剧集该有的
}

func (v *VideoSeriesCreate) Fun() (err error) {
	if v.IDsJson != "" {
		v.IDs, err = ToSliceInt64(v.IDsJson)
	}
	return
}

// 这个可以对其内部进行parentID指定或者指定video里的seriesID
// 跟create一样验证不严格
// 是否进行级联删除
// 也可以进行id的增减
// 可以进行一级字段的改
type VideoSeriesUpdate struct {
	ID          int64   `form:"ID" validate:"required"`
	Delete      string  `form:"Delete" validate:"omitempty,oneof=DeleteCASCADE Delete"`
	Name        string  `form:"Name" validate:"required_without_all=Delete Description IDsJson"`
	Description string  `form:"Description" validate:"required_without_all=Delete Name IDsJson"`
	IDsJson     string  `form:"IDs" validate:"required_without_all=Delete Name Description"`
	IDs         []int64 `form:"-"`
}

func (v *VideoSeriesUpdate) Fun() (err error) {
	if v.IDsJson != "" {
		v.IDs, err = ToSliceInt64(v.IDsJson)
	}
	return
}

type CommentPost struct {
	ParentID int64  `form:"ParentID" validate:"required_with=UserID2"` // 如果是在一个comment下边进行评论当然有这个
	VideoID  int64  `form:"VideoID" validate:"required"`               // 哪个video下
	UserID2  int64  `form:"UserID2" validate:"required_with=ParentID"` // 对谁说
	Comment  string `form:"Comment" validate:"required"`               // 评论啊
}

type CommentSearch struct {
	ID        int64  `form:"ID" validate:"gt=0"`
	Type      string `form:"Type" validate:"oneof=ParentID VideoID UserID"`
	LimitTime int64  `form:"LimitTime"`
	Page      int    `form:"Page" validate:"gte=0"`
	OrderBy   string `form:"OrderBy" validate:"omitempty,oneof=CreatedAt"`
	DESC      bool   `form:"DESC"`
}

type VideoDetle struct {
	ID    int64 `validate:"required"`
	Detle int64 `validate:"required"`
}

type ViewSearch struct {
	ID int64 `form:"ID" validate:"required"`
	// IDs        int64  `form:"IDs" validate:"required"`
	Type      string `form:"Type" validate:"oneof=VideoID UserID"`
	LimitTime int64  `form:"LimitTime"`
	Page      int    `form:"Page" validate:"gte=0"`
	OrderBy   string `form:"OrderBy" validate:"omitempty,oneof=LastViewTime"`
	DESC      bool   `form:"DESC"`
}

type CollectedVideoSearch struct {
	ID        int64  `form:"ID" validate:"required"`
	Type      string `form:"Type" validate:"oneof=UserID VideoID"`
	LimitTime int64  `form:"LimitTime"`
	Page      int    `form:"Page" validate:"gte=0"`
	OrderBy   string `form:"OrderBy" validate:"omitempty,oneof=LastViewTime"`
	DESC      bool   `form:"DESC"`
}
