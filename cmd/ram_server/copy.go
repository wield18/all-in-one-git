package ramserver

import "time"

func Like(column string) string {
	return column + " LIKE ?"
}

func IN(column string) string {
	return column + " IN ?"
}

const (
	// 通用字段
	SourcePlain     = "Source"    // slog一般构建args
	UserTokenPlain  = "UserToken" // 这里必须使用crypto加密的token再跟对应的id进行cipher.AEAD的aes加密,传输时加密解密形成俩个字段
	CaptchaValPlain = "CaptchaVal"
	IPPlain         = "IP"
	UserAgantPlain  = "UserAgant"
	PlatformPlain   = "Platform"
	DeletePlain     = "Delete"
	IDsPlain        = "IDs"
	SearchTypePlain = "SearchType"

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

	// Series
	SeriesPlain    = "Series"
	BaseOnPlain    = "BaseOn"
	IDsAppendPlain = "IDsAppend"
	IDsRemovePlain = "IDsRemove"

	// contextToken
	UserCodePlain = "UserCode"

	// UserV1字段
	UserV1Plain      = "UserV1"
	PasswordPlain    = "Password"
	OldPasswordPlain = "OldPassword"
	EmailPlain       = "Email"
	PhonePlain       = "Phone"
	StatusPlain      = "Status"

	BytesPlain           = "Bytes"
	LoginLocationsPlain  = "LoginLocations"
	VideosPlain          = "Videos"
	CollectedVideosPlain = "CollectedVideos"
	FollowingPlain       = "Following"
	FollowingCountPlain  = "FollowingCount"
	FollowersPlain       = "Followers"
	FollowersCountPlain  = "FollowersCount"

	// LoginLocation字段
	LoginLocationPlain = "LoginLocation"

	// CollectedVideo字段
	CollectedVideoPlain = "CollectedVideo"

	// UserFollow字段
	UserFollowPlain = "UserFollow"
	FollowerIDPlain = "FollowerID"
	FolloweeIDPlain = "FolloweeID"

	// VideoSeries
	VideosCountPlain   = "VideosCount"
	SumViewsCountPlain = "SumViewsCount"

	// Video
	SeriesIDPlain       = "SeriesID"
	VideoPlain          = "Video"
	DurationPlain       = "Duration"
	VideoIDPlain        = "VideoID"
	DescriptionPlain    = "Description"
	ViewsCountPlain     = "ViewsCount"
	ViewsPlain          = "Views"
	TagsPlain           = "Tags"
	CollectedCountPlain = "CollectedCount"
	CollectedsPlain     = "Collecteds"
	CommentsCountPlain  = "CommentsCount"
	CommentsPlain       = "Comments"

	// Comment
	CommentPlain       = "Comment"
	CommentIDPlain     = "CommentID"
	ParentIDPlain      = "ParentID"
	ChildrenPlain      = "Children"
	UserID2Plain       = "UserID2"
	ChildrenCountPlain = "ChildrenCount"
)

const (
	UserTokenDuration      = 7 * 24 * 3600 * time.Second
	RestoreRate            = 0.5
	CaptchaLen             = 6
	RandKeyLen             = 16
	LoginLocationTimeLimit = 7 * 24 * 3600 * time.Second
	PageSize               = 10
)
