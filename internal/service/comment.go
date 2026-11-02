package service

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	consumerouter "all-in-one/internal/consume_router"
	"all-in-one/internal/entity"
	"all-in-one/pkg/consumer"
	httperr "all-in-one/pkg/ekit/http_err"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	commentUserID2NotFound = "commentUserID2NotFound"
	validateWrong          = "validateWrong"
)

func PostComment(mdb *gorm.DB, p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) {
		ierr := postComment(mdb, c, p)
		if !ierr.IsZero() {
			base.AddResultIErrAndArgs(c, ierr)
			c.Abort()
			return
		}
		base.ResultGin(c, "ok")
	}
}

func postComment(mdb *gorm.DB, c *gin.Context, p *consumer.ConsumerPool) httperr.IError {
	contextTokenc := c.MustGet(constant.ContextToken).(entity.ContextToken)
	form := c.MustGet(constant.FormData).(*entity.CommentPost)
	_ = entity.Comment{}
	// 对于有parentID的得检查一下parentID跟其下边的有没有我们当前的对谁说,如果没有必然非法
	if form.ParentID != 0 {
		aMapUserID := map[int64]struct{}{} // 所有当前评论里的id
		data := entity.Comment{}
		err := mdb.Preload(ChildrensPlain, func(db *gorm.DB) *gorm.DB {
			return db.Select(UserIDPlain, UserID2Plain)
		}).Select(UserIDPlain).
			Where(Equal(VideoIDPlain), form.VideoID).Where(Equal(IDPlain), form.ParentID).
			First(&data).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound(notFound(CommentPlain))
			}
			return ErrInternalServerError(err.Error())
		}
		aMapUserID[data.UserID] = struct{}{}
		if len(data.Childrens) != 0 {
			for _, v := range data.Childrens {
				aMapUserID[v.UserID] = struct{}{}
				aMapUserID[v.UserID2] = struct{}{}
			}
		}
		// 必须找到
		if _, ok := aMapUserID[form.UserID2]; !ok {
			return ErrBadRequest(commentUserID2NotFound)
		}

	}
	data := &entity.Comment{
		VideoID: form.VideoID,
		Comment: form.Comment,
		UserID:  contextTokenc.UserID,
	}
	if form.ParentID != 0 {
		data.ParentID = &form.ParentID
		data.UserID2 = form.UserID2
	}

	// 创建并更新父数据
	err := mdb.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(data).Error
		if err != nil {
			return err
		}
		return tx.Model(&entity.Comment{}).Where(Equal(IDPlain), form.ParentID).
			Update(ChildrenCountPlain, gorm.Expr(Subtract(ChildrenCountPlain), -1)).Error
	})
	if err != nil {
		return ErrInternalServerError(err.Error())
	}

	p.Submit(consumerouter.MysqlUpdateVideoCommentCount, consumer.NewConsumeData(entity.VideoDetle{
		ID:    form.VideoID,
		Detle: 1,
	}, func(data interface{}, e httperr.IError) {
		if !e.IsZero() {
			slog.Error(e.Msg, base.NewSlogArgs().SlogJsonDate(data)...)
		}
	}))
	return httperr.IError{}
}

func DeleteComment(mdb *gorm.DB, p *consumer.ConsumerPool) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, deleteComment(mdb, c, p)) }
}

func deleteComment(mdb *gorm.DB, c *gin.Context, p *consumer.ConsumerPool) httperr.IError {
	idStr, ok := c.GetQuery(IDPlain)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if !ok || err != nil {
		return ErrBadRequest(validateWrong)
	}
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)

	var aComment entity.Comment
	err = mdb.Select(VideoIDPlain).Where(Equal(IDPlain), id).
		Where(Equal(UserIDPlain), contextToken.UserID).
		First(&aComment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound(notFound(CommentPlain))
		}
		ErrInternalServerError(err.Error())
	}
	err = mdb.Where(Equal(IDPlain), id).
		Where(Equal(UserIDPlain), contextToken.UserID).
		Delete(&entity.Comment{}).Error
	if err != nil {
		ErrInternalServerError(err.Error())
	}
	p.Submit(consumerouter.MysqlUpdateVideoCommentCount, consumer.NewConsumeData(entity.VideoDetle{
		ID:    aComment.VideoID,
		Detle: -1,
	}, func(data interface{}, e httperr.IError) {
		if !e.IsZero() {
			slog.Error(e.Msg, base.NewSlogArgs().SlogJsonDate(data)...)
		}
	}))
	c.Set(constant.Result, "ok")
	return httperr.IError{}
}

func SearchComments(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, searchComment(c, mdb)) }
}

func searchComment(c *gin.Context, mdb *gorm.DB) httperr.IError {
	query := c.MustGet(constant.QueryData).(*entity.CommentSearch)

	switch query.Type {
	case VideoIDPlain:
		// 只能选择userid2为0的
		mdb = mdb.Where(Equal(query.Type), query.ID).Where(Equal(UserID2Plain), 0)
	case ParentIDPlain:
		mdb = mdb.Where(Equal(query.Type), query.ID)
	case UserIDPlain:
		mdb = mdb.Where(EqualOr(UserIDPlain, UserID2Plain), query.ID, query.ID)
	}

	// 有时间限制就写上
	if query.LimitTime != 0 {
		mdb = mdb.Where(Greater(CreatedAtPlain), time.Now().Add(-(time.Duration(query.LimitTime) * time.Second)))
	}

	mdb = mdb.Offset(query.Page * constant.PageSize).Limit(constant.PageSize)

	suffix := ""
	if query.DESC {
		suffix = " DESC"
	}
	switch query.OrderBy {
	case CreatedAtPlain:
		mdb = mdb.Order(query.OrderBy + suffix)
	}
	datas := []entity.Comment{}
	err := mdb.Find(&datas).Error
	if err != nil {
		return ErrInternalServerError(err.Error())
	}

	res := []map[string]interface{}{}
	for _, v := range datas {
		res = append(res, v.ToMapMany())
	}
	ginData(c, res)
	return httperr.IError{}
}

// func SearchCommentByParentID(mdb *gorm.DB) func(c *gin.Context) {
// 	return func(c *gin.Context) {
// 		idStr, ok := c.GetQuery(IDPlain)
// 		id, err := strconv.ParseInt(idStr, 10, 64)
// 		if !ok || err != nil {
// 			base.AddResultIErrAndArgs(c, ErrBadRequest(validateWrong))
// 			c.Abort()
// 			return
// 		}
// 		res := []map[string]interface{}{}
// 		datas := []entity.Comment{}
// 		err = mdb.Select(ChildrenCountPlain, UserIDPlain, CommentPlain, CreatedAtPlain).
// 			Offset(query.Page * PageSize).Limit(PageSize).
// 			Order(OrderDESC(CreatedAtPlain)).Find(&datas).Error
// 		for _, data := range datas {
// 			res = append(res, data.ToMapForTransmitMany())
// 		}
// 		if err != nil {
// 			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
// 			c.Abort()
// 			return
// 		}
// 		setData(c, res)
// 	}
// }
