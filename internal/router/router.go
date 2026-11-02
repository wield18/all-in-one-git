package router

import (
	"all-in-one/config"
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"all-in-one/internal/entity"
	"all-in-one/internal/local"
	"all-in-one/internal/middleware"
	"all-in-one/internal/service"
	httperr "all-in-one/pkg/ekit/http_err"

	"github.com/gin-gonic/gin"
)

var (
	GetLocation          = "/GetLocation"
	UpdateUserSuper      = "/UpdateUserSuper"
	GetCaptchaEmail      = "/GetCaptchaEmail"
	GetB64sCaptcha       = "/GetB64sCaptcha"
	SearchUserMany       = "/SearchUserMany"
	SearchUserOne        = "/SearchUserOne"
	Logout               = "/Logout"
	Login                = "/Login"
	Register             = "/Register"
	UpdateUser           = "/UpdateUser"
	UpdateUserVideoBytes = "/UpdateUserVideoBytes"
	RandomRecommendVideo = "/RandomRecommendVideo"
	UpdateVideoBytes     = "/UpdateVideoBytes"
	SearchVideoMany      = "/SearchVideoMany"
	SearchVideoOne       = "/SearchVideoOne"
	UploadVideoFile      = "/UploadVideoFile"
	UpdateVideo          = "/UpdateVideo"
	CreateVideoSeries    = "/CreateVideoSeries"
	UpdateVideoSeries    = "/UpdateVideoSeries"
	UserActiveAdd        = "/UserActiveAdd"
	UserActiveGet        = "/UserActiveGet"
	UserViewVideosUpdate = "/UserViewVideosUpdate"
	PostComment          = "/PostComment"
	DeleteComment        = "/DeleteComment"
	FollowOther          = "/FollowOther"
	DeFollowOther        = "/DeFollowOther"
	CollectingVideo      = "/CollectingVideo"
	DecollectingVideo    = "/DecollectingVideo"
	SearchComments       = "/SearchComments"
	SearchViews          = "/SearchViews"
	SearchFollows        = "/SearchFollows"
	SearchCollections    = "/SearchCollections"
	SearchLoginLocation  = "/SearchLoginLocation"
)

func NewRouter(conf config.Server) *gin.Engine {
	router := gin.New()
	router.MaxMultipartMemory = 500 << 20
	// 跌机是恢复
	router.Use(gin.Recovery())
	router.Use(middleware.Cors(conf.FrontAddress))
	router.Use(middleware.Result)
	router.Use(middleware.SetContextIDAndToken)
	register(router)
	return router
}

func register(r *gin.Engine) {
	withLocation := r.Group("", middleware.SetLocation) // pass
	withLocAndUser := withLocation.Group("", middleware.CheckContextUser)

	// common
	{
		// SetCookie ContextLocation
		// 这是没有就拿,那就让网关处理呗,
		r.GET(GetLocation, // pass
			func(c *gin.Context) { middleware.QueryBindAndValidate(c, &entity.LocationInner{}) },
			service.GetLocation,
		)
		// redisStore CaptchaEmail:CipherDB(email) -> answer
		// return CaptchaX:urlBase64(json(entity.Captcha{Code:CipherDB(email)}))
		withLocation.GET(GetCaptchaEmail, service.GetCaptchaEmailTemp(base.QQDialer, base.P)) // TODO: 这里得改回来
		// redisStore CaptchaB64s:randKey -> answer
		// return CaptchaX:urlBase64(json(entity.Captcha{Code:randKey}))
		withLocation.GET(GetB64sCaptcha, service.GetB64sCaptcha(base.B64sCaptcha, base.P))
	}

	{
		// user
		user := service.UserService{}
		{
			withLocation.POST(Login, // pass
				middleware.CheckCaptcha(base.P, false, constant.CaptchaEmailPlain, constant.CaptchaPhonePlain),
				func(c *gin.Context) { middleware.FormBindAndValidate(c, &entity.UserV1Login{}) },
				user.Login(base.Mysql, base.P),
				user.GenerateUserToken(base.P),
			)
			withLocation.POST(Register, // pass
				middleware.CheckCaptcha(base.P, true, constant.CaptchaEmailPlain, constant.CaptchaPhonePlain),
				func(c *gin.Context) { middleware.FormBindAndValidate(c, &entity.UserV1Register{}) },
				user.Register(base.Mysql),
				user.GenerateUserToken(base.P),
			)
			withLocAndUser.POST(UpdateUser, // pass
				middleware.CheckCaptcha(base.P, false, constant.CaptchaEmailPlain, constant.CaptchaPhonePlain, constant.CaptchaB64sPlain),
				func(c *gin.Context) { middleware.FormBindAndValidate(c, &entity.UserV1Update{}) },
				user.UpdateUser(base.Mysql),
			)
			// super方法
			r.POST(UpdateUserSuper, // pass
				// middleware.CheckSuperContextID,
				func(c *gin.Context) { middleware.FormBindAndValidateWithFunc(c, &entity.UserV1UpdateSuper{}) },
				user.UpdateUserSuper(base.Mysql),
			)

			withLocAndUser.GET(UserActiveAdd, // pass
				user.UserActiveAdd(base.P),
			)
			withLocAndUser.GET(UserActiveGet, // pass
				user.UserActiveGet(base.Mysql),
			)
			withLocAndUser.GET(Logout, // pass
				user.Logout(base.P),
			)
			withLocAndUser.GET(SearchUserMany, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidateWithFunc(c, &entity.UserV1SearchMany{}) },
				user.SearchUserMany(base.Mysql),
			)
			withLocAndUser.GET(SearchUserOne, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidate(c, &entity.UserV1Search{}) },
				user.SearchUserOne(base.Mysql))

			withLocAndUser.GET(UpdateUserVideoBytes, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidateWithFunc(c, &entity.UserVideoBytesUpdate{}) },
				user.UpdateUserVideoBytes(base.Mysql),
			)

		}

		video := service.VideoService{}
		{
			r.GET(RandomRecommendVideo, // pass, 不进行loc验证
				video.RandomRecommendVideo(base.Mysql, local.VideoBuf))

			r.GET(SearchVideoOne,
				video.SearchVideoOneES(base.ES),  // pass
				video.SearchVideoOne(base.Mysql), // pass
			)
			withLocAndUser.GET(SearchVideoMany, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidateWithFunc(c, &entity.VideoSearchMany{}) },
				video.SearchVideoManyES(base.ES),
				video.SearchVideoMany(base.Mysql),
			)
			withLocAndUser.POST(UploadVideoFile, // pass
				func(c *gin.Context) { middleware.FormBindAndValidateWithFunc(c, &entity.VideoFileUpload{}) },
				video.UploadVideoFile(base.Mysql),
			)
			withLocAndUser.POST(UpdateVideo, // pass
				func(c *gin.Context) { middleware.FormBindAndValidateWithFunc(c, &entity.VideoUpdate{}) },
				video.UpdateVideo(base.Mysql),
			)
			withLocAndUser.POST(CreateVideoSeries, // pass
				func(c *gin.Context) { middleware.FormBindAndValidateWithFunc(c, &entity.VideoSeriesCreate{}) },
				video.CreateVideoSeries(base.Mysql),
			)
			withLocAndUser.POST(UpdateVideoSeries, // pass
				func(c *gin.Context) { middleware.FormBindAndValidateWithFunc(c, &entity.VideoSeriesUpdate{}) },
				video.UpdateVideoSeries(base.Mysql),
			)

		}
		other := service.OtherService{}
		{
			// View
			withLocAndUser.GET(UserViewVideosUpdate, // pass
				other.UserViewVideosUpdate(base.P),
			)
			withLocAndUser.GET(SearchViews, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidate(c, &entity.ViewSearch{}) },
				other.SearchViews(base.Mysql))
			// Follow
			withLocAndUser.GET(FollowOther, // pass
				other.FollowOther(base.Mysql),
			)
			withLocAndUser.GET(SearchFollows, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidate(c, &entity.UserFollowSearch{}) },
				other.SearchFollows(base.Mysql))
			// CollectingVideo
			withLocAndUser.GET(CollectingVideo, // pass
				other.CollectingVideo(base.Mysql),
			)
			withLocAndUser.GET(SearchCollections, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidate(c, &entity.CollectedVideoSearch{}) },
				other.SearchCollections(base.Mysql))
			// LoginLocation
			withLocAndUser.GET(SearchLoginLocation, // pass
				func(c *gin.Context) { middleware.QueryBindAndValidateWithFunc(c, &entity.LoginLocationSearch{}) },
				other.SearchLoginLocation(base.Mysql))
		}
		// Comment
		{
			withLocAndUser.POST(PostComment, // pass
				func(c *gin.Context) { middleware.FormBindAndValidate(c, &entity.CommentPost{}) },
				service.PostComment(base.Mysql, base.P),
			)
			withLocAndUser.GET(DeleteComment, // pass
				service.DeleteComment(base.Mysql, base.P),
			)
			withLocation.GET(SearchComments, //  pass
				func(c *gin.Context) { middleware.QueryBindAndValidate(c, &entity.CommentSearch{}) },
				service.SearchComments(base.Mysql),
			)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		base.AddResultIErrAndArgs(c, httperr.ErrNotFound("path: "+c.Request.URL.Path+" not found"))
		c.Abort()
	})

}

func testHandle(c *gin.Context) {
	c.Set(constant.Result, "ok")
}
