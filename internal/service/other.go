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
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OtherService struct{}

// 因为这是add,而且是video的一个简单的view
// 这里不需要管理
// 这里主要是user看了个视频,得有浏览记录
// 而本地计数也是他看了才行不能一段时间发俩遍,而对于这个请求的发送
func (o OtherService) UserViewVideosUpdate(p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
		vIDstr, ok := c.GetQuery(IDPlain)
		vID, err := strconv.ParseInt(vIDstr, 10, 64)
		if !ok || err != nil {
			base.AddResultIErrAndArgs(c, ErrBadRequest(lackVideoID))
			c.Abort()
			return
		}
		// 这里不能这样,因为user的观看必须得上传上去
		// 而对于视频的被观看量的统计得走本地统一上传
		ierr := p.BlockSubmit(consumerouter.MysqlUserViewVideos,
			consumer.NewConsumeData(entity.View{
				UserID: contextToken.UserID, VideoID: vID, LastViewTime: time.Now()}, nil))
		if !ierr.IsZero() {
			base.AddResultIErrAndArgs(c, ierr)
			c.Abort()
			return
		}
		// 本地计数,统一上传
		// 这里直接双层map搞定了
		local.ViewCount.Incr(vID, contextToken.UserID)

		base.ResultGin(c, "ok")
	}
}

func (o OtherService) CollectingVideo(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, o.collectingVideo(mdb, c)) }
}

func (o OtherService) collectingVideo(mdb *gorm.DB, c *gin.Context) httperr.IError {
	idStr, ok := c.GetQuery(IDPlain)
	videoID, err := strconv.ParseInt(idStr, 10, 64)
	if !ok || err != nil {
		return ErrBadRequest(notFound(IDPlain))
	}
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)

	res := mdb.Where(Equal(UserIDPlain), contextToken.UserID).
		Where(Equal(VideoIDPlain), videoID).First(&entity.CollectedVideo{})
	if res.RowsAffected == 0 {
		data := entity.CollectedVideo{
			UserID:  contextToken.UserID,
			VideoID: videoID,
		}
		err = mdb.Create(&data).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrBadRequest(alreadyCollectThisOne)
			}
			return ErrInternalServerError(err.Error())
		}
		c.Set(constant.Result, "ok: 收藏")
	} else {
		err = mdb.Where(Equal(UserIDPlain), contextToken.UserID).Where(Equal(VideoIDPlain), videoID).
			Delete(&entity.CollectedVideo{}).Error
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
		c.Set(constant.Result, "ok: 取消收藏")
	}

	return httperr.IError{}
}

func (o OtherService) FollowOther(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, o.followOther(mdb, c)) }
}

// follow 跟 disfolloe 搁一起吧
// 尝试查一下,有就删除,没有就创建
func (o OtherService) followOther(mdb *gorm.DB, c *gin.Context) httperr.IError {
	idStr, ok := c.GetQuery(IDPlain)
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if !ok || err != nil {
		return ErrBadRequest(notFound(IDPlain))
	}
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
	if contextToken.UserID == userID {
		return ErrBadRequest(noSelfFollow)
	}
	res := mdb.Where(Equal(FollowerIDPlain), contextToken.UserID).Where(Equal(FolloweeIDPlain), userID).
		First(&entity.UserFollow{})
	if res.RowsAffected == 0 {
		// 没找到创建
		data := entity.UserFollow{
			FollowerID: contextToken.UserID,
			FolloweeID: userID,
		}
		err = mdb.Create(&data).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrBadRequest(alreadyFollowThisOne)
			}
			return ErrInternalServerError(err.Error())
		}
		c.Set(constant.Result, "ok: 关注")
	} else {
		err = mdb.Where(Equal(FollowerIDPlain), contextToken.UserID).Where(Equal(FolloweeIDPlain), userID).
			Delete(&entity.UserFollow{}).Error
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
		c.Set(constant.Result, "ok: 取消关注")
	}

	return httperr.IError{}

}

func (o OtherService) SearchViews(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, o.searchViews(c, mdb)) }
}
func (o OtherService) searchViews(c *gin.Context, mdb *gorm.DB) httperr.IError {
	query := c.MustGet(constant.QueryData).(*entity.ViewSearch)
	mdb = mdb.Where(Equal(query.Type), query.ID)
	datas := []entity.View{}
	err := repository.Paging(mdb, query.LimitTime, query.Page, query.OrderBy, query.DESC).Find(&datas).Error
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	dats := []map[string]any{}
	for _, v := range datas {
		dats = append(dats, v.ToMap())
	}
	ginData(c, dats)
	return httperr.IError{}
}

func (o OtherService) SearchFollows(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, o.searchFollows(c, mdb)) }
}
func (o OtherService) searchFollows(c *gin.Context, mdb *gorm.DB) httperr.IError {
	query := c.MustGet(constant.QueryData).(*entity.UserFollowSearch)
	mdb = mdb.Where(Equal(query.Type), query.ID)
	datas := []entity.UserFollow{}
	err := repository.Paging(mdb, query.LimitTime, query.Page, query.OrderBy, query.DESC).Find(&datas).Error
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	ginData(c, datas)
	return httperr.IError{}
}

func (o OtherService) SearchCollections(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, o.searchCollections(c, mdb)) }
}
func (o OtherService) searchCollections(c *gin.Context, mdb *gorm.DB) httperr.IError {
	query := c.MustGet(constant.QueryData).(*entity.CollectedVideoSearch)
	mdb = mdb.Where(Equal(query.Type), query.ID)
	datas := []entity.CollectedVideo{}
	err := repository.Paging(mdb, query.LimitTime, query.Page, query.OrderBy, query.DESC).Find(&datas).Error
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	ginData(c, datas)
	return httperr.IError{}
}

func (o OtherService) SearchLoginLocation(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, o.searchLoginLocation(c, mdb)) }
}

func (o OtherService) searchLoginLocation(c *gin.Context, mdb *gorm.DB) httperr.IError {
	query := c.MustGet(constant.QueryData).(*entity.LoginLocationSearch)
	mdb = mdb.Where(Equal(query.Type), query.Key)
	datas := []entity.LoginLocation{}
	err := repository.Paging(mdb, query.LimitTime, query.Page, query.OrderBy, query.DESC).Find(&datas).Error
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	ginData(c, datas)
	return httperr.IError{}
}
