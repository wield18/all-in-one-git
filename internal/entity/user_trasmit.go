package entity

import (
	"strconv"
)

var timelimitIsZero = "timelimit is zero"

type UserV1Register struct {
	Name     string `form:"Name" validate:"required"`
	Password string `form:"Password" validate:"required"`
}

type UserV1Login struct {
	Name     string `form:"Name" validate:"required_without_all=Phone Email"`
	Password string `form:"Password" validate:"required"`
	Phone    string `form:"Phone" validate:"required_without_all=Email Name"`
	Email    string `form:"Email" validate:"required_without_all=Phone Name"`
}

type UserV1Update struct {
	Name     string `form:"Name" validate:"omitempty,noPunct"`
	Password string `form:"Password"`
	// 使用 Captcha 获得
	// Phone    string `validate:"omitempty,e164" exactlyone:"1of6" form:"Phone"`
	// Email    string `validate:"omitempty,email" exactlyone:"1of6" form:"Email"`
	Status string `form:"Status"`
	Delete bool   `form:"Delete"`

	OldPassword string `form:"OldPassword" validate:"required_with=Password"`
}

// 这里只要根据id进行改就行
type UserV1UpdateSuper struct {
	ID         int64  `form:"ID" validate:"required"`
	Name       string `exactlyone:"1of6" form:"Name" validate:"omitempty,noPunct"`
	Password   string `exactlyone:"1of6" form:"Password"`
	Phone      string `validate:"omitempty,e164" exactlyone:"1of6" form:"Phone"`
	Email      string `validate:"omitempty,email" exactlyone:"1of6" form:"Email"`
	Status     string `exactlyone:"1of6" form:"Status"`
	Bytes      []byte
	ByteString string `exactlyone:"1of6" form:"Bytes"`
	Delete     bool   `exactlyone:"1of6" form:"Delete"`
}

func (u *UserV1UpdateSuper) Fun() (err error) {
	return u.DecodeHttp()
}

func (u *UserV1UpdateSuper) DecodeHttp() (err error) {
	if u.ByteString != "" {
		u.Bytes = []byte(u.ByteString)
	}
	return
}

type UserV1Search struct {
	ID    int64  `form:"ID"`
	Name  string `form:"Name"`
	Phone string `form:"Phone" validate:"omitempty,e164"`
	Email string `form:"Email" validate:"omitempty,email"`
}

// 可以根据ids查多个用户，或者单id来查followee跟follower
// 或者Name进行模糊查询
type UserV1SearchMany struct {
	ID      int64   `form:"ID" exactlyone:"1of5"`
	Type    string  `form:"Type" validate:"oneof=Followee Follower Name"`
	IDs     []int64 `form:"-"`
	IDsJson string  `form:"IDs" exactlyone:"1of5"`
	Name    string  `form:"Name" exactlyone:"1of5" validate:"omitempty,noPunct"`

	LimitTime int64  `form:"LimitTime"`
	Page      int    `form:"Page" validate:"gte=0"`
	OrderBy   string `form:"OrderBy" validate:"omitempty,oneof=Name CreatedAt"`
	DESC      bool   `form:"DESC"`
}

func (u *UserV1SearchMany) Fun() (err error) {
	if u.IDsJson != "" {
		u.IDs, err = ToSliceInt64(u.IDsJson)
	}
	return err
}

type UserVideoBytesUpdate struct {
	UserID           int64    `form:"UserID" validate:"required_without_all=TagsJson ExcludedTagsJson"`
	Tags             []string `form:"_"`
	TagsJson         string   `form:"Tags" validate:"required_without_all=UserID ExcludedTagsJson"`
	ExcludedTags     []string `form:"_"`
	ExcludedTagsJson string   `form:"ExcludedTags" validate:"required_without_all=UserID TagsJson"`
}

func (u *UserVideoBytesUpdate) Fun() (err error) {
	if u.TagsJson != "" {
		u.Tags, err = ToSliceString(u.TagsJson)
	}
	if u.ExcludedTagsJson != "" {
		u.ExcludedTags, err = ToSliceString(u.ExcludedTagsJson)
	}
	return
}

type UserFollowSearch struct {
	ID int64 `form:"ID" validate:"required"`
	// IDs        int64  `form:"IDs" validate:"required"`
	Type      string `form:"Type" validate:"oneof=FollowerID FolloweeID"`
	LimitTime int64  `form:"LimitTime"`
	Page      int    `form:"Page" validate:"gte=0"`
	OrderBy   string `form:"OrderBy" validate:"omitempty,oneof=CreatedAt"`
	DESC      bool   `form:"DESC"`
}

type LoginLocationSearch struct {
	Key       string `form:"Key" validate:"required"`
	Type      string `form:"Type" validate:"oneof=UserID Location"`
	LimitTime int64  `form:"LimitTime"`
	Page      int    `form:"Page" validate:"gte=0"`
	OrderBy   string `form:"OrderBy" validate:"omitempty,oneof=CreatedAt"`
	DESC      bool   `form:"DESC"`
}

func (l *LoginLocationSearch) Fun() (err error) {
	if l.Type == UserIDPlain {
		_, err = strconv.ParseInt(l.Key, 10, 64)
	}
	return err
}

type ActiveSearch struct {
	Key       string `form:"Key"`
	Type      string `form:"Type" validate:"oneof=UserID Days"`
	LimitTime int64  `form:"LimitTime"`
	Page      int    `form:"Page" validate:"gte=0"`
	OrderBy   string `form:"OrderBy" validate:"omitempty,oneof=CreatedAt"`
	DESC      bool   `form:"DESC"`
}

func (l *ActiveSearch) Fun() (err error) {
	if l.Type == UserIDPlain {
		_, err = strconv.ParseInt(l.Key, 10, 64)
	}
	return err
}
