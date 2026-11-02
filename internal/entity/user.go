package entity

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"encoding/json"
	"sort"
	"time"
)

type UserV1 struct {
	ID                  int64  `gorm:"column:ID;primaryKey;autoIncrement"`
	Name                string `gorm:"column:Name;type:varchar(32);not null;uniqueIndex"` // 字符串，加唯一
	PasswordCrypto      string `gorm:"column:PasswordCrypto;type:char(60);not null"`      // 字符串
	PasswordPlainString string `gorm:"-"`

	// 因为 base.CipherDB 每次加密的值不一样
	PhoneCipherDB    string `gorm:"column:PhoneCipherDB;type:varchar(255);comment:手机号(含国家前缀)"` // base.CipherDB
	PhoneHash        []byte `gorm:"column:PhoneHash;type:binary(32);uniqueIndex"`              // HMACHashString 哈希列供查询
	PhonePlainString string `gorm:"-"`

	EmailCipherDB    string `gorm:"column:EmailCipherDB;type:varchar(255);"`      // 邮箱 唯一 base.CipherDB
	EmailHash        []byte `gorm:"column:EmailHash;type:binary(32);uniqueIndex"` // HMACHashString 哈希列供查询
	EmailPlainString string `gorm:"-"`

	Status string `gorm:"column:Status;type:enum('happy','sad','anxious','angry','calm');default:'happy'"` // 枚举
	// 二进制数据,本想存点map[string]interface{}的 有用了
	Bytes []byte `gorm:"column:Bytes;type:json;comment:JSON数组存点想给自己的标签"` // 二进制数据

	CreatedAt time.Time `gorm:"column:CreatedAt;autoCreateTime"` // 自动创建时间

	// 这里保留的外表连接主要还是一件事,会不会preload,会不会一起更新,一般是不会,只需要外键关联就行
	FollowingCount int64 `gorm:"column:FollowingCount;default:0"` // 我关注的人数
	FollowerCount  int64 `gorm:"column:FollowerCount;default:0"`  // 我的粉丝数
	// 我关注的人（我是 FollowerID）
	// Following []UserFollow `gorm:"foreignKey:FollowerID;references:ID;constraint:OnDelete:CASCADE"`
	// 关注我的人 / 我的粉丝（我是 FolloweeID）
	// Followers []UserFollow `gorm:"foreignKey:FolloweeID;references:ID;constraint:OnDelete:CASCADE"`

	LoginLocations []LoginLocation `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Actives        []Active        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	// Videos         []Video         `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:SET NULL;"`

	// Views      []View `gorm:"foreignKey:VideoID;references:ID;constraint:OnDelete:SET NULL"`
	// CollectedVideos      []CollectedVideo `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	// Comment []Comment `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:SET NULL;"`
	// 这是我收藏的video
	// CollectedVideos []CollectedVideo `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

type Bytes struct {
	Video VideoBytes
}

func (b *Bytes) ToJsonByte() []byte {
	by, _ := json.Marshal(b)
	return by
}

type VideoBytes struct {
	IncludedTags    []VideoBytesTag // 默认拿到手是order过的,最长十个大概
	ExcludedUserIDs []int64         // 最长可能5个吧
}

func (v *VideoBytes) AddExcludedUserIDs(ids ...int64) {
	v.ExcludedUserIDs = append(v.ExcludedUserIDs, ids...)
	if len(v.ExcludedUserIDs) > 5 {
		v.ExcludedUserIDs = v.ExcludedUserIDs[len(v.ExcludedUserIDs)-5:]
	}
}

func (v *VideoBytes) IncreaseIncludedTags(tags ...string) {
	if len(tags) == 0 {
		return
	}
	// 空值初始
	if len(v.IncludedTags) == 0 {
		share := 1 / float64(len(tags))
		tmp := make([]VideoBytesTag, len(tags))
		for i, v := range tags {
			tmp[i] = VideoBytesTag{Tag: v, Ratio: share}
		}
		v.IncludedTags = tmp
		return
	}
	// 这里可能有新增的,也可能是旧的在里边的

	aMap := map[string]struct{}{} // aMap 供tag查询
	for _, v := range tags {
		aMap[v] = struct{}{}
	}
	all := float64(0)                      // 可以分配的百分比
	foundTags := map[string]struct{}{}     // 真正原来就有的tag
	allFoundOrNewTags := []VideoBytesTag{} // 找到或最后新增的
	allLeftTags := []VideoBytesTag{}       // 没命中的tag但份额足够
	for _, tagS := range v.IncludedTags {
		// 对于命中的
		if _, ok := aMap[tagS.Tag]; ok {
			foundTags[tagS.Tag] = struct{}{}
			allFoundOrNewTags = append(allFoundOrNewTags, tagS)
			continue
		}
		// 对于太小的直接
		if tagS.Ratio < 0.05 {
			all += tagS.Ratio
			continue
		}
		// 可以但不是传进来的部分
		tagS.Ratio = tagS.Ratio / 2
		all += tagS.Ratio
		allLeftTags = append(allLeftTags, tagS)
	}
	// 如果到最后还有些传入的没找到
	for _, tag := range tags {
		if _, ok := foundTags[tag]; !ok {
			allFoundOrNewTags = append(allFoundOrNewTags, VideoBytesTag{Tag: tag})
		}
	}

	// 对于找到根新的进行更新
	share := all / float64(len(tags))
	for i := 0; i < len(allFoundOrNewTags); i++ {
		allFoundOrNewTags[i].Ratio += share
	}
	// 最后更新
	v.IncludedTags = append(allLeftTags, allFoundOrNewTags...)
	// 排序倒序
	sort.Slice(v.IncludedTags, func(i, j int) bool { return v.IncludedTags[i].Ratio > v.IncludedTags[j].Ratio })

}

func (v *VideoBytes) DeleteTags(tags ...string) {
	if len(tags) == 0 {
		return
	}
	aMap := map[string]struct{}{} // aMap 供tag查询
	for _, v := range tags {
		aMap[v] = struct{}{}
	}
	leftTags := []VideoBytesTag{}
	all := float64(0)
	for _, v := range v.IncludedTags {
		if _, ok := aMap[v.Tag]; !ok {
			leftTags = append(leftTags, v)
			continue
		}
		all += v.Ratio
	}
	share := all / float64(len(leftTags))
	for i := 0; i < len(leftTags); i++ {
		leftTags[i].Ratio += share
	}
	// 这里不用排序
	v.IncludedTags = leftTags
}

type VideoBytesTag struct {
	Tag   string
	Ratio float64
}

func (UserV1) TableName() string {
	return "UserV1"
}

func (u *UserV1) UnmarshalBytes() *Bytes {
	abc := Bytes{}
	json.Unmarshal(u.Bytes, &abc)
	return &abc
}

// 这一般都是创建时使用,或更新吧
func (u *UserV1) EncodeDB() *UserV1 {
	if u.PasswordPlainString != "" {
		u.PasswordCrypto, _ = base.HashCryptoCode(u.PasswordPlainString)
	}
	if u.PhonePlainString != "" {
		u.PhoneCipherDB = base.CipherDB.Encrypt(u.PhonePlainString)
		u.PhoneHash = base.HashHMACString(u.PhonePlainString)
	}
	if u.EmailPlainString != "" {
		u.EmailCipherDB = base.CipherDB.Encrypt(u.EmailPlainString)
		u.EmailHash = base.HashHMACString(u.EmailPlainString)
	}
	if len(u.LoginLocations) == 1 { // 这里认为不可能会有同时多个location在创建或更新
		u.LoginLocations[0].CodeCrypto, _ = base.HashCryptoCode(u.LoginLocations[0].CodePlainText)
	}
	return u
}

func (u *UserV1) EncodeDBForSearch() *UserV1 {
	if u.PhonePlainString != "" {
		u.PhoneHash = base.HashHMACString(u.PhonePlainString)
	}
	if u.EmailPlainString != "" {
		u.EmailHash = base.HashHMACString(u.EmailPlainString)
	}
	if len(u.LoginLocations) == 1 { // 这里认为不可能会有同时多个location在创建或更新
		u.LoginLocations[0].CodeCrypto, _ = base.HashCryptoCode(u.LoginLocations[0].CodePlainText)
	}
	return u
}

func (u *UserV1) DecodeDB() {
	if u.EmailCipherDB != "" {
		u.EmailPlainString, _ = base.CipherDB.Decrypt(u.EmailCipherDB)
	}
	if u.PhoneCipherDB != "" {
		u.PhonePlainString, _ = base.CipherDB.Decrypt(u.PhoneCipherDB)
	}
}

// 返回使用
// 不进行关联表的加载,也不返回关联表数据
// 真要自己在关联表写
func (u *UserV1) ToMap() map[string]interface{} {
	aMap := map[string]interface{}{}
	if u.ID != 0 {
		aMap[constant.IDPlain] = u.ID
	}
	if u.Name != "" {
		aMap[constant.NamePlain] = u.Name
	}
	if u.PhoneCipherDB != "" {
		plainText, _ := base.CipherDB.Decrypt(u.PhoneCipherDB)
		aMap[constant.PhonePlain] = plainText
	}
	if u.EmailCipherDB != "" {
		plainText, _ := base.CipherDB.Decrypt(u.EmailCipherDB)
		aMap[constant.EmailPlain] = plainText
	}
	if u.Status != "" {
		aMap[constant.StatusPlain] = u.Status
	}
	if !u.CreatedAt.IsZero() {
		aMap[constant.CreatedAtPlain] = u.CreatedAt
	}
	aMap[constant.FollowersCountPlain] = u.FollowerCount
	aMap[constant.FollowingCountPlain] = u.FollowingCount
	aMap[ActivesPlain] = u.Actives
	aMap[LoginLocationsPlain] = u.LoginLocations
	return aMap
}

func (u *UserV1) ToMapMany() map[string]interface{} {
	aMap := map[string]interface{}{}
	if u.ID != 0 {
		aMap[constant.IDPlain] = u.ID
	}
	if u.Name != "" {
		aMap[constant.NamePlain] = u.Name
	}
	if u.Status != "" {
		aMap[constant.StatusPlain] = u.Status
	}
	if !u.CreatedAt.IsZero() {
		aMap[constant.CreatedAtPlain] = u.CreatedAt
	}
	aMap[constant.FollowersCountPlain] = u.FollowerCount
	aMap[constant.FollowingCountPlain] = u.FollowingCount
	return aMap
}

// UserFollow 关注关系：FollowerID 关注 FolloweeID
type UserFollow struct {
	ID int64 `gorm:"column:ID;primaryKey;autoIncrement"`

	FollowerID int64     `gorm:"column:FollowerID;not null;uniqueIndex:uk_follow"` // 关注者（粉丝）
	FolloweeID int64     `gorm:"column:FolloweeID;not null;uniqueIndex:uk_follow"` // 被关注者
	CreatedAt  time.Time `gorm:"column:CreatedAt;autoCreateTime"`
	// 可选：关系状态、分组、备注等
	// Status string `gorm:"column:Status;type:enum('active','blocked');default:'active'"`
}

func (UserFollow) TableName() string {
	return "UserFollow"
}

type CollectedVideo struct {
	ID int64 `gorm:"column:ID;primaryKey;autoIncrement"`
	// UserID,LoginTime 加入联合索引
	UserID    int64     `gorm:"column:UserID;not null;index;comment:用户ID"`
	VideoID   int64     `gorm:"column:VideoID;not null;index;comment:视频ID"`
	CreatedAt time.Time `gorm:"column:CreatedAt;autoCreateTime"`
}

func (CollectedVideo) TableName() string {
	return "CollectedVideo"
}

// 这里估计以后使用save?我不清楚一般来说不可能在 Location 或 CodeCrypto 加唯一,所以基本不可能重复
// 还是直接create完事 所以 autoCreateTime
type LoginLocation struct {
	ID int64 `gorm:"column:ID;primaryKey;autoIncrement" json:"-"`
	// UserID,LoginTime 加入联合索引
	UserID int64 `gorm:"column:UserID;not null;index;comment:用户ID" validate:"required" json:"-"`
	// 真实存储位置
	Location string `gorm:"column:Location;type:varchar(255)" validate:"required"`
	// 用啥登录,哪个平台,这个只检查是否是新设备,就行,主要地址
	// 我可能就想省事而已就使用了 crypto 吧
	CodeCrypto    string    `gorm:"column:CodeCrypto;type:char(60);not null" json:"-"`
	CodePlainText string    `gorm:"-" validate:"required" json:"-"`
	LoginTime     time.Time `gorm:"column:LoginTime;autoCreateTime"`
}

func (v *LoginLocation) EncodeDB() *LoginLocation {
	if v.CodePlainText != "" {
		v.CodeCrypto, _ = base.HashCryptoCode(v.CodePlainText)
	}
	return v
}

func (LoginLocation) TableName() string {
	return "LoginLocation"
}

// 活跃时间一般是短时token吧,就比如说4小时,一天分6段,一个小时发一次,只要没有就发,获得token,有没有token都不影响
// 只是记录一下你的活跃// 这个估计得让网关来发获得,检查cookie,获得cookie,访问后端记录活跃
// 这个应该不用直接关联主表因为我如果要根据UserID来查这个也可能吧,基本不可能,直接访问这个表完事
type Active struct {
	ID int64 `gorm:"column:ID;primaryKey;autoIncrement" json:"-"`
	// 这里有复合索引
	UserID int64     `gorm:"column:UserID;not null;index;comment:用户ID" json:"-"`
	Days   time.Time `gorm:"column:Days;type:date"`

	DaysActiveCount int64     `gorm:"column:DaysActiveCount"`
	LastActiveTime  time.Time `gorm:"column:LastActiveTime"`
}

func (Active) TableName() string {
	return "Active"
}
