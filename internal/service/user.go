package service

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	consumerouter "all-in-one/internal/consume_router"
	"all-in-one/internal/entity"
	"all-in-one/internal/local"
	"all-in-one/internal/repository"
	"all-in-one/pkg/consumer"
	httperr "all-in-one/pkg/ekit/http_err"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	userNotFound           = "userNotFound"
	passwordNotEqual       = "passwordNotEqual"
	needCaptchaValidate    = "needCaptchaValidate"
	captchaIsNotChecked    = "captchaIsNotChecked"
	updateNotMatch         = "updateNotMatch"
	searchWhereOptionNil   = "searchWhereOptionNil"
	asyncSubmitTokenErr    = "asyncSubmitTokenErr"
	nope                   = "nope"
	logoutRedisDeleteWrong = "logoutRedisDeleteWrong"
	searchTypeNoMatched    = "searchTypeNoMatched"
	lackVideoID            = "lackVideoID"
	alreadyFollowThisOne   = "alreadyFollowThisOne"
	alreadyCollectThisOne  = "alreadyCollectThisOne"
	noSelfFollow           = "noSelfFollow"
	noData                 = "noData"
)

type UserService struct{}

func notFound(column string) string {
	return column + "NotFound"
}

// 这个查询只是简单查询100%命中的
func (u UserService) SearchUserOne(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := c.MustGet(constant.QueryData).(*entity.UserV1Search)
		if query.ID == 0 && query.Name == "" && query.Email == "" && query.Phone == "" {
			query.ID = c.MustGet(constant.ContextToken).(entity.ContextToken).UserID
		}
		dataUser := &entity.UserV1{
			ID:               query.ID,
			Name:             query.Name,
			EmailPlainString: query.Email,
			PhonePlainString: query.Phone,
		}

		var result *gorm.DB
		var data entity.UserV1
		result = mdb.Preload(LoginLocationsPlain, func(db *gorm.DB) *gorm.DB {
			return db.Select(UserIDPlain, LocationPlain).Limit(1).
				Order(clause.OrderByColumn{Column: clause.Column{Name: LoginTimePlain}, Desc: true})
		}).Preload(ActivesPlain, func(db *gorm.DB) *gorm.DB {
			return db.Select(UserIDPlain, DaysPlain, DaysActiveCountPlain, LastActiveTimePlain).Limit(1).
				Order(clause.OrderByColumn{Column: clause.Column{Name: DaysPlain}, Desc: true})
		}).
			Select(IDPlain, NamePlain, StatusPlain, PhoneCipherDBPlain, EmailCipherDBPlain,
				FollowingCountPlain, FollowerCountPlain, CreatedAtPlain).
			Where(dataUser.EncodeDBForSearch()).First(&data)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				base.AddResultIErrAndArgs(c, ErrNotFound(userNotFound))
			} else {
				base.AddResultIErrAndArgs(c, ErrInternalServerError(result.Error.Error()))
			}
			c.Abort()
			return
		}
		c.Set(constant.Result, map[string]interface{}{
			constant.Data: data.ToMap(),
		})
	}
}

func (u UserService) SearchUserMany(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, u.searchUserMany(c, mdb)) }
}

func (u UserService) searchUserMany(c *gin.Context, mdb *gorm.DB) httperr.IError {
	query := c.MustGet(constant.QueryData).(*entity.UserV1SearchMany)

	if query.Name == "" && query.ID != 0 { // 得不从name查才可能进行ids的初始
		datas := []entity.UserFollow{}
		switch query.Type {
		case FolloweePlain: // 查的是被关注的人,啥概念,这个id是关注者吗?
			mdb = mdb.Select(FolloweeIDPlain).Where(Equal(FollowerIDPlain), query.ID)
		case FollowerPlain:
			mdb = mdb.Select(FollowerIDPlain).Where(Equal(FolloweeIDPlain), query.ID)
		}
		res := mdb.Offset(query.Page * constant.PageSize).Limit(constant.PageSize).Find(&datas)
		if res.Error != nil {
			return ErrInternalServerError(res.Error.Error())
		}
		if res.RowsAffected == 0 {
			c.Set(constant.Result, noFound)
			return ErrNotFound(noFound)
		}
		ids := []int64{}
		switch query.Type {
		case FolloweePlain:
			for _, v := range datas {
				ids = append(ids, v.FolloweeID)
			}
		case FollowerPlain:
			for _, v := range datas {
				ids = append(ids, v.FollowerID)
			}
		}
		query.IDs = ids
	}

	mdb = mdb.Session(&gorm.Session{NewDB: true}).Select(IDPlain, NamePlain, StatusPlain, CreatedAtPlain, FollowingCountPlain, FollowerCountPlain)
	if query.Name != "" {
		mdb = mdb.Where(Like(NamePlain), query.Name)
	} else {
		mdb = mdb.Where(IN(IDPlain), query.IDs)
	}

	users := []entity.UserV1{}
	res := repository.Paging(mdb, query.LimitTime, query.Page, query.OrderBy, query.DESC).Find(&users)
	if res.Error != nil {
		return ErrInternalServerError(res.Error.Error())
	}
	if res.RowsAffected == 0 {
		c.Set(constant.Result, noFound)
		return httperr.IError{}
	}
	results := []map[string]interface{}{}
	for _, v := range users {
		results = append(results, v.ToMapMany())
	}

	ginData(c, results)
	return httperr.IError{}
}

// SearchType string `validate:"required_with=ID,oneof=Following Followers LoginLocations Videos CollectedVideos"`
// func (u UserService) SearchUserSecondTable(mdb *gorm.DB) func(c *gin.Context) {
// 	return func(c *gin.Context) { ginAbortIerr(c, u.searchUserSecondTable(c, mdb)) }
// }

// // LoginLocation, Active, UserFollow(Following, Followers), Video, View, CollectedVideo, Comment, Series
// func (u UserService) searchUserSecondTable(c *gin.Context, mdb *gorm.DB) (ierr httperr.IError) {
// 	query := c.MustGet(constant.QueryData).(*entity.UserV1SearchSecondTable)
// 	var results []map[string]interface{}
// 	userID := c.MustGet(constant.ContextToken).(entity.ContextToken).UserID
// 	var err error
// 	// 这里用interface来接一下data
// 	switch query.SearchType {
// 	case LoginLocationsPlain:
// 		data := []entity.LoginLocation{}
// 		model := &entity.LoginLocation{}
// 		mdb = mdb.Model(model).Select(LocationPlain, LoginTimePlain)
// 		if query.UserID == 0 { // 这里可以指定查自己或别人
// 			query.UserID = userID
// 		}
// 		mdb = mdb.Where(Equal(UserIDPlain), query.UserID)

// 		err = mdb.Where(Greater(LoginTimePlain), time.Now().Add(-constant.LoginLocationTimeLimit)).
// 			Order(OrderDESC(LoginTimePlain)).
// 			Offset(query.Page * PageSize).Limit(PageSize).Find(&data).Error
// 		for _, v := range data {
// 			results = append(results, v.ToMapForTransmitMany())
// 		}
// 	case VideosPlain:
// 		// 视频记录
// 		data := []entity.Video{}
// 		model := &entity.Video{}

// 		mdb = mdb.Model(model).Select(IDPlain, NamePlain, CreatedAtPlain, PathPlain, ViewsCountPlain, TagsPlain, CollectedVideosCountPlain, CommentsCountPlain)
// 		if query.UserID == 0 { // 这里可以指定查自己或别人
// 			mdb = mdb.Where(Equal(UserIDPlain),
// 				c.MustGet(constant.ContextToken).(entity.ContextToken).UserID)
// 		} else {
// 			mdb = mdb.Where(Equal(UserIDPlain), query.UserID)
// 		}

// 		err = mdb.Order(OrderDESC(CreatedAtPlain)).
// 			Offset(query.Page * PageSize).Limit(PageSize).Find(&data).Error
// 		for _, v := range data {
// 			results = append(results, v.ToDataMany())
// 		}
// 	case CollectedVideosPlain:
// 		// 收藏记录
// 		// 得在 CollectedVideo 进行子查询
// 		// 这里使用子查询原因在于需要通过CollectedVideo来找video返回
// 		data := []entity.Video{}
// 		model := &entity.Video{}

// 		subQuery := mdb.Table(CollectedVideoPlain).Select(VideoIDPlain)
// 		if query.UserID == 0 { // 这里可以指定查自己或别人
// 			subQuery = subQuery.Where(Equal(UserIDPlain),
// 				c.MustGet(constant.ContextToken).(entity.ContextToken).UserID)
// 		} else {
// 			subQuery = subQuery.Where(Equal(UserIDPlain), query.UserID)
// 		}

// 		err = mdb.Model(model).Select(IDPlain, NamePlain, CreatedAtPlain, PathPlain, ViewsCountPlain, TagsPlain, CollectedVideosCountPlain, CommentsCountPlain).
// 			Where(IN(IDPlain), subQuery).Order(OrderDESC(CreatedAtPlain)).
// 			Offset(query.Page * PageSize).Limit(PageSize).Find(&data).Error
// 		for _, v := range data {
// 			results = append(results, v.ToDataMany())
// 		}
// 	case FollowingPlain:
// 		// 关注记录
// 		data := []entity.UserV1{}
// 		model := &entity.UserV1{}
// 		subQuery := mdb.Table(UserFollowPlain).
// 			Select(FolloweeIDPlain)
// 		if query.UserID == 0 { // 这里可以指定查自己或别人
// 			subQuery = subQuery.Where(Equal(FollowerIDPlain),
// 				c.MustGet(constant.ContextToken).(entity.ContextToken).UserID)
// 		} else {
// 			subQuery = subQuery.Where(Equal(FollowerIDPlain), query.UserID)
// 		}

// 		err = mdb.Model(model).Select(IDPlain, NamePlain, StatusPlain, FollowingCountPlain, FollowerCountPlain, CreatedAtPlain).
// 			Where(IN(IDPlain), subQuery).
// 			Offset(query.Page * PageSize).Limit(PageSize).Find(&data).Error
// 		for _, v := range data {
// 			results = append(results, v.ToMapForTransmitMany())
// 		}

// 	case FollowersPlain:
// 		// 被关注记录
// 		data := []entity.UserV1{}
// 		model := &entity.UserV1{}
// 		subQuery := mdb.Table(UserFollowPlain).
// 			Select(FollowerIDPlain)
// 		if query.UserID == 0 { // 这里可以指定查自己或别人
// 			subQuery = subQuery.Where(Equal(FolloweeIDPlain),
// 				c.MustGet(constant.ContextToken).(entity.ContextToken).UserID)
// 		} else {
// 			subQuery = subQuery.Where(Equal(FolloweeIDPlain), query.UserID)
// 		}
// 		err = mdb.Model(model).Select(IDPlain, NamePlain, StatusPlain, FollowingCountPlain, FollowerCountPlain, CreatedAtPlain).
// 			Where(IN(IDPlain), subQuery).
// 			Offset(query.Page * PageSize).Limit(PageSize).Find(&data).Error
// 		for _, v := range data {
// 			results = append(results, v.ToMapForTransmitMany())
// 		}
// 	}
// 	if len(results) == 0 {
// 		base.AddResultIErrAndArgs(c, ErrNotFound(nope))
// 		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
// 		}
// 		c.Abort()
// 		return
// 	}
// 	c.Set(constant.Result, map[string]interface{}{
// 		constant.Data: results,
// 	})
// 	return
// }

func (u UserService) Logout(p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
		key := base.Sprintf(constant.UserToken, contextToken.UserCode)
		p.Submit(consumerouter.RedisKeyDel,
			consumer.NewConsumeData(key, func(data interface{}, e httperr.IError) {
				if !e.IsZero() {
					slog.Error(logoutRedisDeleteWrong, base.NewSlogArgs(2).SlogJsonDate(data)...)
					return
				}
			}))
		local.TokenStorage.Del(key)
		c.SetCookie(constant.ContextToken, "nil", -1, "", "", false, true)
		c.Set(constant.Result, "ok")
	}
}
func (u UserService) Login(mdb *gorm.DB, p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		ginAbortIerr(c, u.login(c, mdb, p))
	}
}

func (u UserService) login(c *gin.Context, mdb *gorm.DB, p *consumer.ConsumerPool) httperr.IError {
	aForm := c.MustGet(constant.FormData).(*entity.UserV1Login)
	aLoc := c.MustGet(constant.ContextLocation).(entity.ContextLocation)
	aUser := &entity.UserV1{}
	dataUser := &entity.UserV1{
		Name:             aForm.Name,
		EmailPlainString: aForm.Email,
		PhonePlainString: aForm.Phone,
	}
	model := &entity.UserV1{}
	//
	result := mdb.Model(model).Preload(
		LoginLocationsPlain, func(db *gorm.DB) *gorm.DB {
			return db.Select(UserIDPlain, LocationPlain, CodeCryptoPlain).Where(Greater(LoginTimePlain), time.Now().Add(-constant.LoginLocationTimeLimit))
		},
	).
		Select(
			IDPlain,
			PasswordCryptoPlain,
			PhoneCipherDBPlain,
			EmailCipherDBPlain,
		).Where(dataUser.EncodeDBForSearch()).First(aUser)
	if result.RowsAffected == 0 { // 没查到呗
		return ErrNotFound(userNotFound)
	}
	// 检查密码
	if !base.CheckCryptoCode(aUser.PasswordCrypto, aForm.Password) {
		return ErrBadRequest(passwordNotEqual)
	}
	// 地点位置
	var hasLoc bool
	for _, v := range aUser.LoginLocations {
		if base.CheckCryptoCode(v.CodeCrypto, aLoc.LocationPlain.SprintfSortStrings()) {
			hasLoc = true
		}
	}
	isCheckCaptcha := c.GetBool(constant.IsCaptchaChecked)
	if !hasLoc && !isCheckCaptcha { // 新登录又没有进行captcha的检查
		aUser.DecodeDB()
		// 返回200但请再次尝试
		resolution := make(map[string]string)
		if aUser.PhonePlainString != "" {
			resolution[constant.PhonePlain] = aUser.PhonePlainString
		}
		if aUser.EmailPlainString != "" {
			resolution[constant.EmailPlain] = aUser.EmailPlainString
		}
		c.Set(constant.Result, map[string]interface{}{constant.ResulotionPlain: resolution})
		return base.IErrWarn(needCaptchaValidate)
	}
	// 这里说明经过了captcha的检查
	// 必须得在没有loc的情况下进行更新loc
	if !hasLoc {
		// 更新location就行然后继续,这里直接发个任务就行
		p.Submit(consumerouter.MysqlCreateLoginLocation, consumer.NewConsumeData(entity.LoginLocation{
			UserID:        aUser.ID,
			Location:      aLoc.Location,
			CodePlainText: aLoc.LocationPlain.SprintfSortStrings(),
		}, func(data interface{}, e httperr.IError) {
			if !e.IsZero() {
				slog.Error(e.Msg, base.NewSlogArgs(2).SlogSource().SlogJsonDate(data)...)
			}
		}))
	}

	// 剩下的只有有loc跟captcha的 或 有loc但没captcha的
	// 都是直接进行下个,认为正常登录
	c.Set(constant.UserID, aUser.ID)
	return httperr.IError{}
}

// 这里如果想生成必须得传个id过来
// 这里得redis上传还有本地存
func (u UserService) GenerateUserToken(p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.GetInt64(constant.UserID)
		code, _ := base.HashCryptoCode(strconv.FormatInt(id, 10))
		key := base.Sprintf(constant.UserToken, code)
		// redis
		p.Submit(consumerouter.RedisKeyValueSet, consumer.NewConsumeData(entity.KeyValueRedis{
			Key:     key,
			Timeout: constant.UserTokenDuration,
		}, func(data interface{}, e httperr.IError) {
			if !e.IsZero() {
				slog.Error(asyncSubmitTokenErr, base.NewSlogArgs(2).SlogJsonDate(data)...)
			}
		}))

		// 本地
		local.TokenStorage.Store(code, time.Now().Add(constant.UserTokenDuration))
		// 这里得使用urlBase(json(ContextToken)) 才行
		c.SetCookie(constant.ContextToken, entity.ToUrlSafeBase64String(entity.ContextToken{UserID: id, UserCode: code}), int(constant.UserTokenDuration), "", "", false, true)
		c.Set(constant.Result, "ok")
	}
}

func (u UserService) Register(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		aForm := c.MustGet(constant.FormData).(*entity.UserV1Register)
		aLoc := c.MustGet(constant.ContextLocation).(entity.ContextLocation)
		aCapt := c.MustGet(constant.Captcha).(*entity.Captcha)
		// 已经进行完了captcha检查
		// fieldStr 字段使用, 得进行 CipherDB 的解码
		fieldStr, err := base.CipherDB.Decrypt(aCapt.Code)
		if err != nil {
			c.Abort()
			base.AddResultIErrAndArgs(c, ErrBadRequest(err.Error()))
			return
		}

		aUser := &entity.UserV1{
			Name:                aForm.Name,
			PasswordPlainString: aForm.Password,
			LoginLocations: []entity.LoginLocation{
				{Location: aLoc.Location, CodePlainText: base.SprintfSortStrings(aLoc.LocationPlain.IP, aLoc.LocationPlain.UserAgant, aLoc.LocationPlain.Platform)},
			},
		}
		// 写一下captcha的
		if aCapt.Type == CaptchaEmailPlain {
			aUser.EmailPlainString = fieldStr
		} else {
			aUser.PhonePlainString = fieldStr
		}
		err = mdb.Create(aUser.EncodeDB()).Error
		if err != nil {
			c.Abort()
			if repository.IsErrDuplicated(err) {
				base.AddResultIErrAndArgs(c, ErrConflict(err.Error()))
				return
			}
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			return
		}
		c.Set(constant.UserID, aUser.ID)
	}

}

func (u UserService) UpdateUser(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
		aForm := c.MustGet(constant.FormData).(*entity.UserV1Update)
		if aForm.Delete {
			err := mdb.Delete(&entity.UserV1{}, contextToken.UserID).Error
			if err != nil {
				base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			}
			c.SetCookie(constant.ContextToken, "nil", -1, "", "", false, true)
			c.Set(constant.Result, "ok")
			return
		}
		dataUser, ierr := u.getDataUser(c, mdb, aForm, contextToken)
		if !ierr.IsZero() {
			base.AddResultIErrAndArgs(c, ierr)
			c.Abort()
			return
		}
		err := mdb.Model(&entity.UserV1{}).Where(Equal(IDPlain), contextToken.UserID).Updates(dataUser.EncodeDB()).Error
		if err != nil {
			c.Abort()
			if repository.IsErrDuplicated(err) {
				base.AddResultIErrAndArgs(c, ErrConflict(err.Error()))
				return
			}
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			return
		}
		// 这说明是得删除
		c.Set(constant.Result, "ok")
	}
}

func (u UserService) getDataUser(c *gin.Context, mdb *gorm.DB, aForm *entity.UserV1Update, contextToken entity.ContextToken) (dataUser *entity.UserV1, ierr httperr.IError) {
	// 前边已经进行了只能更新一个的限制,还要对应的captcha的检查,
	dataUser = &entity.UserV1{}
	if aForm.Name != "" {
		dataUser.Name = aForm.Name
		return
	}
	if aForm.Status != "" {
		dataUser.Status = aForm.Status
		return
	}
	captcha := c.MustGet(constant.Captcha).(*entity.Captcha)
	// 下边三个都有对应的captcha
	if aForm.Password != "" {
		if captcha.Type != constant.CaptchaB64sPlain {
			str := base.Sprintf(captchaIsNotChecked, captcha.Type, constant.CaptchaB64sPlain)
			return nil, ErrBadRequest(str)
		}
		// 得进行旧密码的先前检查
		if !repository.CheckPassword(mdb, contextToken.UserID, aForm.OldPassword) {
			return nil, ErrBadRequest(passwordNotEqual)
		}
		dataUser.PasswordPlainString = aForm.Password
		return
	}
	if captcha.IsChecked == true {
		data, err := base.CipherDB.Decrypt(captcha.Code)
		if err != nil {
			return nil, ErrBadRequest(err.Error())
		}
		switch captcha.Type {
		case constant.CaptchaPhonePlain:
			dataUser.EmailPlainString = data
			return
		case constant.CaptchaEmailPlain:
			dataUser.PhonePlainString = data
			return
		}
	}
	return nil, ErrBadRequest(noData)
}

func (u UserService) UpdateUserSuper(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var result *gorm.DB
		aForm := c.MustGet(constant.FormData).(*entity.UserV1UpdateSuper)
		model := &entity.UserV1{}
		if aForm.Delete {
			result = mdb.Delete(model, aForm.ID)
		} else {
			data := &entity.UserV1{
				Name:                aForm.Name,
				PasswordPlainString: aForm.Password,
				PhonePlainString:    aForm.Phone,
				EmailPlainString:    aForm.Email,
				Status:              aForm.Status,
				Bytes:               aForm.Bytes,
			}
			result = mdb.Model(model).Where(Equal(IDPlain), aForm.ID).Updates(data.EncodeDB())
		}
		// 报错一定没有更新,不更新可能没报错

		// 报错
		if result.Error != nil {
			if repository.IsErrDuplicated(result.Error) {
				base.AddResultIErrAndArgs(c, ErrConflict(result.Error.Error()))
				return
			}
			base.AddResultIErrAndArgs(c, ErrInternalServerError(updateNotMatch))
			return
		}
		// 没有任何更新
		if result.RowsAffected == 0 {
			c.Abort()
			base.AddResultIErrAndArgs(c, ErrNotFound(updateNotMatch))
			return
		}

		c.Set(constant.Result, "ok")
	}
}

func (u UserService) UserActiveAdd(p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
		t := time.Now()
		y1, m1, d1 := t.Date()
		days := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
		// 这只是创建当前天的 active
		// 主要是重复更新lastActiveTime
		data := entity.Active{UserID: contextToken.UserID, Days: days, DaysActiveCount: 0, LastActiveTime: t}
		ierr := p.BlockSubmit(consumerouter.MysqlUserActive, consumer.NewConsumeData(
			&data, nil,
		))
		if !ierr.IsZero() {
			base.AddResultIErrAndArgs(c, ierr)
			c.Abort()
			return
		}
		// 这里进行 activeCount 的自增
		local.ActiveCount.Incr(data.ID)
		c.Set(constant.Result, "ok")
	}
}

func (u UserService) UserActiveGet(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)

		// 这里只是获得最后一天的active
		data := entity.Active{}
		err := mdb.Select(DaysPlain, DaysActiveCountPlain, LastActiveTimePlain).Where(Equal(UserIDPlain), contextToken.UserID).First(&data).Error
		if err != nil {
			c.Abort()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				base.AddResultIErrAndArgs(c, ErrNotFound(err.Error()))
				return
			}
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			return
		}
		// 这里只是拿到一个时间段的,不能说明是这一天
		incr := local.ActiveCount.GetIncr(contextToken.UserID)

		aMap := map[string]interface{}{
			DaysPlain:            data.Days,
			DaysActiveCountPlain: data.DaysActiveCount,
			LastActiveTimePlain:  data.LastActiveTime,
		}
		if base.IsToday(data.Days) {
			aMap[DaysActiveCountPlain] = data.DaysActiveCount + incr
		}

		c.Set(constant.Result, aMap)
	}
}

func (u UserService) UpdateUserVideoBytes(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, u.updateUserVideoBytes(c, mdb)) }
}

func (u UserService) updateUserVideoBytes(c *gin.Context, mdb *gorm.DB) (ierr httperr.IError) {
	data := c.MustGet(constant.QueryData).(*entity.UserVideoBytesUpdate)
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
	dbData := entity.UserV1{}
	err := mdb.Select(BytesPlain).First(&dbData, contextToken.UserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBadRequest(err.Error())
		}
		return ErrInternalServerError(err.Error())
	}
	bytes := entity.Bytes{}
	if len(dbData.Bytes) != 0 {
		err = json.Unmarshal(dbData.Bytes, &bytes)
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
	}
	if data.UserID != 0 {
		bytes.Video.AddExcludedUserIDs(data.UserID)
	}
	bytes.Video.IncreaseIncludedTags(data.Tags...)
	bytes.Video.DeleteTags(data.ExcludedTags...)

	err = mdb.Model(&entity.UserV1{}).
		Where(Equal(IDPlain), contextToken.UserID).
		Update(BytesPlain, bytes.ToJsonByte()).Error
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	c.Set(constant.Result, "ok")
	return
}
