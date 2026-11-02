package service

import (
	"strings"
	"time"
)

// ---------------entity copy

func Like(column string) string {
	return column + " LIKE ?"
}

func IN(column string) string {
	return column + " IN (?)"
}

func Equal(column string) string {
	return column + " = ?"
}

func EqualOr(column string, columns ...string) string {
	conditions := []string{}
	conditions = append(conditions, column+" = ?")
	if len(columns) != 0 {
		for _, v := range columns {
			conditions = append(conditions, v+" = ?")
		}
		sb := strings.Builder{}
		sb.WriteString("(" + conditions[0])
		for i := 1; i < len(conditions); i++ {
			sb.WriteString(" or " + conditions[i])
		}
		sb.WriteString(")")

		return sb.String()
	}
	return conditions[0]
}

func IsNull(column string) string {
	return column + " is null"
}

func NotEqual(column string) string {
	return column + " != ?"
}

func Less(column string) string {
	return column + " < ?"
}

func Greater(column string) string {
	return column + " > ?"
}

// Where("JSON_CONTAINS(Tags, ?)", `"gaojian"`).   // 注意双引号
func JSON_CONTAINS(column string) string {
	return "JSON_CONTAINS(" + column + ", ?)"
}

func OrderDESC(column string) string {
	return column + " DESC"
}

func Subtract(column string) string {
	return column + " - ?"
}

func Add(column string) string {
	return column + " - ?"
}

const (
	// 通用字段
	SourcePlain     = "Source"    // slog一般构建args
	UserTokenPlain  = "UserToken" // 这里必须使用crypto加密的token再跟对应的id进行cipher.AEAD的aes加密,传输时加密解密形成俩个字段
	CaptchaValPlain = "CaptchaVal"
	IPPlain         = "IP"
	UserAgantPlain  = "UserAgant"
	PlatformPlain   = "Platform"
	ChildrensPlain  = "Childrens"

	// cookie
	CookiePlain    = "Cookie"
	CookieValPlain = "CookieVal"
	TimeoutPlain   = "Timeout"
	MaxAgePlain    = "MaxAge"

	KeyPlain     = "Key"
	TypePlain    = "Type"
	RawPathPlain = "RawPath"
	// 这仨一般进行类型命中
	CpatchaTypePlain  = "CpatchaType"
	CaptchaB64sPlain  = "CaptchaB64s"
	CaptchaEmailPlain = "CaptchaEmail"
	CaptchaPhonePlain = "CaptchaPhone"
	ResulotionPlain   = "Resulotion"

	UserIDPlain    = "UserID"
	IDPlain        = "ID"
	NamePlain      = "Name"
	CreatedAtPlain = "CreatedAt"
	PathPlain      = "Path"

	// contextToken
	UserCodePlain = "UserCode"

	// UserV1字段
	UserV1Plain         = "UserV1"
	EmailPlain          = "Email"
	PhonePlain          = "Phone"
	StatusPlain         = "Status"
	PasswordCryptoPlain = "PasswordCrypto"
	PhoneCipherDBPlain  = "PhoneCipherDB"
	EmailCipherDBPlain  = "EmailCipherDB"

	BytesPlain           = "Bytes"
	LoginLocationsPlain  = "LoginLocations"
	VideosPlain          = "Videos"
	CollectedVideosPlain = "CollectedVideos"
	FollowingPlain       = "Following"
	FollowingCountPlain  = "FollowingCount"
	FollowersPlain       = "Followers"
	FollowerCountPlain   = "FollowerCount"
	ActivesPlain         = "Actives"
	// LoginLocation字段
	LoginLocationPlain = "LoginLocation"
	LoginTimePlain     = "LoginTime"
	LocationPlain      = "Location"
	CodeCryptoPlain    = "CodeCrypto"

	// CollectedVideo字段
	CollectedVideoPlain = "CollectedVideo"

	// UserFollow字段
	UserFollowPlain = "UserFollow"
	FollowerIDPlain = "FollowerID"
	FolloweeIDPlain = "FolloweeID"
	FollowerPlain   = "Follower"
	FolloweePlain   = "Followee"

	// VideoSeries
	VideosCountPlain = "VideosCount"
	LayerPlain       = "Layer"
	BaseOnPlain      = "BaseOn"

	// Video
	LastViewTimePlain         = "LastViewTime"
	SeriesPlain               = "Series"
	SeriesNamePlain           = "SeriesName"
	SeriesIDPlain             = "SeriesID"
	VideoPlain                = "Video"
	DurationPlain             = "Duration"
	VideoIDPlain              = "VideoID"
	DescriptionPlain          = "Description"
	ViewsCountPlain           = "ViewsCount"
	ViewsPlain                = "Views"
	TagsPlain                 = "Tags"
	CollectedCountPlain       = "CollectedCount"
	CollectedsPlain           = "Collecteds"
	CommentsCountPlain        = "CommentsCount"
	CommentsPlain             = "Comments"
	CollectedVideosCountPlain = "CollectedVideosCount"

	// Comment
	CommentIDplain     = "CommentID"
	CommentPlain       = "Comment"
	ParentIDPlain      = "ParentID"
	ChildrenPlain      = "Children"
	UserID2Plain       = "UserID2"
	ChildrenCountPlain = "ChildrenCount"

	// Active
	DaysPlain            = "Days"
	DaysActiveCountPlain = "DaysActiveCount"
	LastActiveTimePlain  = "LastActiveTime"
)

const (
	UserTokenDuration      = 7 * 24 * 3600 * time.Second
	RestoreRate            = 0.5
	CaptchaLen             = 6
	RandKeyLen             = 16
	LoginLocationTimeLimit = 7 * 24 * 3600 * time.Second
	PageSize               = 10
)
