package constant

import "time"

// 加ptr的一般存储指针,没加的一般就是entity的对应命名的结构体
// 加plain的就是字符串本身

// 一般来说传入是啥名就是绑定啥名

const (
	Result          = "Result"       // 绑定结构体
	ContextID       = "ContextID"    // 绑定String
	ContextToken    = "ContextToken" // 绑定结构体,
	UserToken       = "UserToken"    // 作为redis的键使用,主要是 ContextToken 里边的 Code 字段使用了 HashCryptoCode(userID)
	ContextLocation = "LocationX"    // // 绑定结构体 这里的变量名对应着entity的ContextLocation,这里也是

	QueryData = "QueryData" // 绑定结构体指针 // 这个主要在于c.Set()使用的绑定结构体的键名
	JsonData  = "JSONData"  // 绑定结构体指针 // 这个主要在于c.Set()使用的绑定结构体的键名
	FormData  = "FormData"  // 绑定结构体指针 // 这个主要在于c.Set()使用的绑定结构体的键名

	IErrAndArgs      = "SlogIErrAndArgs"  // 绑定[]*constant.IErrAndArgs // 最后result使用的err查询,通常进行slog跟最终的返回处理
	Captcha          = "CaptchaX"         // 绑定结构体指针 一般返回或者检查请求的query的captcha字段使用,这是tokon验证字段
	IsCaptchaChecked = "IsCaptchaChecked" // // 绑定bool,没检查一般直接就是false
	CaptchaImage     = "Captcha"          // 返回b64s使用
	UserID           = "UserID"           // 传输使用生成userToken
	Data             = "Data"             // 返回result时使用map[string]interface{}的时候的数据键名
)

const (
	SourcePlain     = "Source" // slog一般构建args
	CaptchaValPlain = "CaptchaVal"
	IPPlain         = "IP"
	UserAgantPlain  = "UserAgant"
	PlatformPlain   = "Platform"

	KeyPlain     = "Key"
	TypePlain    = "Type"
	RawPathPlain = "RawPath"
	// 这仨一般进行类型命中
	CpatchaTypePlain  = "CpatchaType"
	CaptchaB64sPlain  = "CaptchaB64s"
	CaptchaEmailPlain = "CaptchaEmail"
	CaptchaPhonePlain = "CaptchaPhone"
	ResulotionPlain   = "Resulotion"

	UserIDPlain = "UserID"

	UserCodePlain = "UserCode"

	// UserV1字段
	IDPlain              = "ID"
	NamePlain            = "Name"
	EmailPlain           = "Email"
	PhonePlain           = "Phone"
	StatusPlain          = "Status"
	CreatedAtPlain       = "CreatedAt"
	BytesPlain           = "Bytes"
	LoginLocationsPlain  = "LoginLocations"
	VideosPlain          = "Videos"
	CollectedVideosPlain = "CollectedVideos"
	FollowingPlain       = "Following"
	FollowingCountPlain  = "FollowingCount"
	FollowersPlain       = "Followers"
	FollowersCountPlain  = "FollowersCount"

	// Video
	VideoIDPlain = "VideoID"

	// table名
	UserV1Plain         = "UserV1"
	LoginLocationPlain  = "LoginLocation"
	CollectedVideoPlain = "CollectedVideo"
)

const (
	ContextLocationMaxAge      = 7 * 24 * 3600
	UserTokenDuration          = 7 * 24 * 3600 * time.Second
	UserTokenDurationThreshold = 2 * 24 * 3600 * time.Second
	RestoreRate                = 0.5
	CaptchaLen                 = 6
	RandKeyLen                 = 16
	LoginLocationTimeLimit     = 7 * 24 * 3600 * time.Second
	PageSize                   = 10
	DBSearchMaxIter            = 3
	DBCreateInBatchSize        = 10
	CountTickerGap             = time.Hour * 4
)

const ()
