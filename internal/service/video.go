package service

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	"all-in-one/internal/entity"
	videobuf "all-in-one/internal/local/video_buf"
	"all-in-one/internal/repository"
	httperr "all-in-one/pkg/ekit/http_err"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	idWrong                          = "idWrong"
	openFileFailed                   = "open file failed"
	readFileFailed                   = "read file failed"
	sha256Mismatch                   = "sha256 mismatch"
	iterToMany                       = "iterToMany"
	attempToUpdateOtherVideo         = "attempToUpdateOtherVideo"
	attempToSeiresVideoThatHasSeries = "attempToSeiresVideoThatHasSeries"
	mdbVideoRecordNotFound           = "mdbVideoRecordNotFound"
	searchNotMatch                   = "searchNotMatch"
	seriesLayerWrong                 = "seriesLayerWrong"
)

type VideoService struct{}

func (v VideoService) SearchVideoManyES(es *elasticsearch.TypedClient) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, v.searchVideoManyES(c, es)) }
}

func (v VideoService) searchVideoManyES(c *gin.Context, es *elasticsearch.TypedClient) httperr.IError {
	query := c.MustGet(constant.QueryData).(*entity.VideoSearchMany)
	must := map[string]any{}
	switch query.Type {
	case SeriesIDPlain, UserIDPlain:
		must["term"] = map[string]any{query.Type: query.Key}
	case TagsPlain:
		must["terms"] = map[string]any{query.Type: query.Tags}
	case NamePlain:
		must["title"] = map[string]any{NamePlain: query.Key}
	}
	body, limitFilter := repository.PagingES(map[string]any{}, query.LimitTime, query.Page, query.OrderBy, query.DESC)

	queryBool := map[string]any{}
	if len(must) != 0 {
		queryBool["must"] = []map[string]any{must}
	}
	if len(limitFilter) != 0 {
		queryBool["filter"] = []map[string]any{limitFilter}
	}
	body["query"] = map[string]any{
		"bool": queryBool,
	}
	by, _ := json.Marshal(body)

	model := entity.VideoES{}
	res, err := es.Search().Index(model.IndexName()).Raw(strings.NewReader(string(by))).Do(context.TODO())
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	if len(res.Hits.Hits) != 0 {
		datas := []json.RawMessage{}
		for i := range res.Hits.Hits {
			datas = append(datas, res.Hits.Hits[i].Source_)

		}
		// 查到得主动
		c.Abort()
		ginData(c, datas)
		return httperr.IError{}
	}
	return httperr.IError{}

}

// SeriesIDPlain, IDPlain, UserIDPlain, DurationPlain
// NamePlain, PathPlain
func (v VideoService) SearchVideoMany(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		query := c.MustGet(constant.QueryData).(*entity.VideoSearchMany)
		mdb = mdb.Session(&gorm.Session{NewDB: true}).Select(SeriesIDPlain, IDPlain, UserIDPlain, DurationPlain, NamePlain, PathPlain)

		switch query.Type {
		case SeriesIDPlain, UserIDPlain:
			mdb = mdb.Where(Equal(query.Type), query.Key)
		case TagsPlain:
			l := len(query.Tags)
			if l != 0 {
				if l == 1 {
					mdb = mdb.Where(JSON_CONTAINS(TagsPlain), fmt.Sprintf(`"%s"`, query.Tags[0]))
				}
				// Where("JSON_CONTAINS(Tags, ?)", `["gaojian", "guyuchun"]`)
				mdb = mdb.Where(JSON_CONTAINS(TagsPlain), fmt.Sprintf(`%s`, query.Key))
			}
		case NamePlain:
			mdb = mdb.Where(Like(NamePlain), query.Key)
		}

		datas := []entity.Video{}
		err := repository.Paging(mdb, query.LimitTime, query.Page, query.OrderBy, query.DESC).Find(&datas).Error
		if err != nil {
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			c.Abort()
			return
		}

		// 整理数据
		resDatas := make([]map[string]interface{}, len(datas))
		for i, data := range datas {
			resDatas[i] = data.ToDataMany()
		}
		c.Set(constant.Result, map[string]interface{}{
			constant.Data: resDatas,
		})
	}
}

// video要必中拿id吧
func (v VideoService) SearchVideoOneES(es *elasticsearch.TypedClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		idstr, ok := c.GetQuery(IDPlain)
		_, err := strconv.ParseInt(idstr, 10, 64)
		if !ok || (err != nil) {
			base.AddResultIErrAndArgs(c, ErrBadRequest(idWrong))
			c.Abort()
			return
		}
		data := &entity.VideoES{}
		res, err := es.Get(data.IndexName(), idstr).Do(context.TODO())
		if err != nil {
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			c.Abort()
			return
		}
		if res.Found {
			c.Set(constant.Result, map[string]interface{}{
				constant.Data: res.Source_,
			})
			c.Abort()
		}
	}
}

// video要必中拿id吧
func (v VideoService) SearchVideoOne(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		idstr, ok := c.GetQuery(IDPlain)
		id, err := strconv.ParseInt(idstr, 10, 64)
		if !ok || (err != nil) {
			base.AddResultIErrAndArgs(c, ErrBadRequest(idWrong))
			c.Abort()
			return
		}
		data := &entity.Video{}
		result := mdb.First(data, id)
		if result.RowsAffected == 0 || result.Error != nil {
			base.AddResultIErrAndArgs(c, ErrNotFound(mdbVideoRecordNotFound))
			c.Abort()
			return
		}
		aMap := data.ToDataOne()
		if data.SeriesID != nil {
			aSeries := &entity.Series{}
			res := mdb.Select(NamePlain).First(aSeries, data.SeriesID)
			if res.RowsAffected == 0 || res.Error != nil {
				base.AddResultIErrAndArgs(c, ErrNotFound(mdbVideoRecordNotFound))
				c.Abort()
				return
			}
			aMap[SeriesNamePlain] = aSeries.Name
		}

		c.Set(constant.Result, map[string]interface{}{
			constant.Data: aMap,
		})
	}
}

func (v VideoService) UploadVideoFile(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		req := c.MustGet(constant.FormData).(*entity.VideoFileUpload)
		var err error

		contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)

		var ids []int64
		var seriesID int64
		// 先查一下有没有对应series
		if req.SeriesID != 0 {
			// 尝试查一下series,级联查询
			ids, err = repository.SearchForParentSeriesIDs(mdb, &req.SeriesID)
			if err != nil {
				base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
				c.Abort()
				return
			}
			seriesID = req.SeriesID
		}

		// 储存文件
		var serverHash string
		path := base.VideoStorage.GetFilePath()
		// 这东西只能读一遍
		f, err := req.File.Open()
		if err != nil {
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			return
		}
		defer f.Close()
		// 写入本地
		if req.File.Size < 1<<20 {
			data, err := io.ReadAll(f)
			if err != nil {
				base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
				return
			}
			sum := sha256.Sum256(data)
			serverHash = hex.EncodeToString(sum[:])
			if serverHash != req.SHA256 {
				base.AddResultIErrAndArgs(c, ErrBadRequest(sha256Mismatch))
				c.Abort()
				return
			}
			// 这样的话这data合法,直接写入盘就行
			err = os.WriteFile(path, data, 8644)
			if err != nil {
				base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
				c.Abort()
				return
			}
		} else {
			h := sha256.New()
			// 假设你要落盘，用 TeeReader 同时写盘 + 算 hash
			dst, _ := os.Create(path)
			defer dst.Close()
			tee := io.TeeReader(f, dst)                // 读 f 的同时写入 dst
			if _, err := io.Copy(h, tee); err != nil { // 复制到 hash，触发读
				// copy failed
				base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
				c.Abort()
				return
			}
			serverHash = hex.EncodeToString(h.Sum(nil))
			dst.Close()
			if serverHash != req.SHA256 {
				os.Remove(path) // 校验失败删掉
				base.AddResultIErrAndArgs(c, ErrBadRequest(sha256Mismatch))
				c.Abort()
				return
			}
		}

		// mp4获得时长
		duration, err := base.JustDurationSeconds(path)
		if err != nil {
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			c.Abort()
			return
		}

		// 创建data,写库
		data := &entity.Video{
			UserID:      contextToken.UserID,
			SeriesID:    &seriesID,
			Name:        req.Name,
			Duration:    duration,
			Description: req.Description,
			Path:        path,
			CreatedAt:   time.Now(),
			Tags:        []byte(req.TagsJson),
		}
		err = mdb.Transaction(func(tx *gorm.DB) error {
			// 保存当下
			err = tx.Save(data).Error
			if err != nil {
				return err
			}
			if len(ids) != 0 {
				return tx.Model(&entity.Series{}).Where(IN(IDPlain), ids).
					Update(VideosCountPlain, gorm.Expr(Subtract(VideosCountPlain), -1)).Error
			}
			return err
		})
		if err != nil {
			base.AddResultIErrAndArgs(c, ErrInternalServerError(err.Error()))
			c.Abort()
			return
		}
		c.Set(constant.Result, "ok")
	}
}

func (v VideoService) UpdateVideo(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, v.updateVideo(c, mdb)) }
}

func (v VideoService) updateVideo(c *gin.Context, mdb *gorm.DB) httperr.IError {
	var err error
	form := c.MustGet(constant.FormData).(*entity.VideoUpdate)
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
	// 这里的videoId对应的用户必须是当前contextToken里的
	// 直接找video
	data := entity.Video{}
	err = mdb.Select(UserIDPlain, SeriesIDPlain, ViewsCountPlain).First(&data, form.ID).Error
	if err != nil {
		return ErrInternalServerError(err.Error())
	}

	// 对比下
	if data.UserID != contextToken.UserID {
		return ErrForbidden(attempToUpdateOtherVideo)
	}

	if form.Delete {
		var ids []int64
		// 查一下上边的Series部分
		if data.SeriesID != nil {
			ids, err = repository.SearchForParentSeriesIDs(mdb, data.SeriesID)
			if err != nil {
				return ErrInternalServerError(err.Error())
			}
		}
		err = mdb.Transaction(func(tx *gorm.DB) error {
			err = tx.Delete(&entity.Video{}, form.ID).Error
			if err != nil {
				return err
			}
			if len(ids) != 0 {
				return tx.Model(&entity.Series{}).Where(IN(IDPlain), ids).
					Update(VideosCountPlain, gorm.Expr(Subtract(VideosCountPlain), 1)).
					Update(ViewsCountPlain, gorm.Expr(Subtract(ViewsCountPlain), data.ViewsCount)).Error
			}
			return err
		})
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
	} else {
		mdb = mdb.Model(&entity.Video{}).
			Select(NamePlain, DescriptionPlain, TagsPlain).Where(Equal(IDPlain), form.ID)
		if form.Name != "" {
			mdb = mdb.Update(NamePlain, form.Name)
		}
		if form.Description != "" {
			mdb = mdb.Update(DescriptionPlain, form.Description)
		}
		if form.TagsJson != "" {
			mdb = mdb.Update(TagsPlain, []byte(form.TagsJson))
		}
		err = mdb.Error
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
	}
	c.Set(constant.Result, "ok")
	return httperr.IError{}
}

// VideoSeriesCreate
func (v VideoService) CreateVideoSeries(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, v.createVideoSeries(c, mdb)) }
}

func (v VideoService) createVideoSeries(c *gin.Context, mdb *gorm.DB) httperr.IError {
	form := c.MustGet(constant.FormData).(*entity.VideoSeriesCreate)
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)

	var err error

	// 这里主要是在series里创建一个数据
	parentSeries := entity.Series{
		UserID:      contextToken.UserID,
		Name:        form.Name,
		Description: form.Description,
	}

	// where直接查,看数量对不对
	// 获得关联字段值
	if form.BaseOn == SeriesPlain {
		dataUpdates := []entity.Series{}
		res := mdb.Model(&entity.Series{}).Select(LayerPlain, VideosCountPlain, ViewsCountPlain).
			Where(IN(IDPlain), form.IDs).Where(Equal(UserIDPlain), contextToken.UserID).
			Where(IsNull(ParentIDPlain)).
			Find(&dataUpdates)
		if res.Error != nil {
			return ErrInternalServerError(res.Error.Error())
		}
		if res.RowsAffected == 0 || len(dataUpdates) != len(form.IDs) { // 如果不等大概率是非法说白了
			return ErrForbidden(attempToUpdateOtherVideo)
		}
		// 这里如果说是对于Series的layer大小不一致,一般来说请求参数错误
		currentLayer := dataUpdates[0].Layer
		for i := 1; i < len(dataUpdates); i++ {
			if dataUpdates[i].Layer != currentLayer {
				return ErrBadRequest(seriesLayerWrong)
			}
		}
		parentSeries.Layer = currentLayer + 1
		for _, v := range dataUpdates {
			parentSeries.VideosCount += v.VideosCount
			parentSeries.ViewsCount += v.ViewsCount
		}
	} else {
		dataUpdates := []entity.Video{}
		res := mdb.Model(&entity.Video{}).Select(ViewsCountPlain).
			Where(IN(IDPlain), form.IDs).Where(Equal(UserIDPlain), contextToken.UserID).
			Where(IsNull(SeriesIDPlain)).
			Find(&dataUpdates)
		if res.Error != nil {
			return ErrInternalServerError(res.Error.Error())
		}
		if res.RowsAffected == 0 || len(dataUpdates) != len(form.IDs) { // 如果不等大概率是非法说白了
			return ErrForbidden(attempToUpdateOtherVideo)
		}
		parentSeries.VideosCount = res.RowsAffected
		for _, v := range dataUpdates {
			parentSeries.ViewsCount += v.ViewsCount
		}

	}
	err = mdb.Transaction(func(tx *gorm.DB) error {
		// 创建
		err := tx.Create(&parentSeries).Error
		if err != nil {
			return err
		}
		// 更新关联表
		if form.BaseOn == SeriesPlain {
			return tx.Model(&entity.Series{}).Where(IN(IDPlain), form.IDs).Update(ParentIDPlain, parentSeries.ID).Error
		}
		return tx.Model(&entity.Video{}).Where(IN(IDPlain), form.IDs).Update(SeriesIDPlain, parentSeries.ID).Error

	})
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	c.Set(constant.Result, "ok")
	return httperr.IError{}

}

func (v VideoService) UpdateVideoSeries(mdb *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) { ginAbortIerr(c, v.updateVideoSeries(c, mdb)) }
}

func (v VideoService) updateVideoSeries(c *gin.Context, mdb *gorm.DB) (ierr httperr.IError) {
	// 这个无论咋样,基于哪更新哪
	// 还得把上层更新了,当然最多迭代三次
	form := c.MustGet(constant.FormData).(*entity.VideoSeriesUpdate)
	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)
	var err error

	data := entity.Series{}
	err = mdb.Where(Equal(UserIDPlain), contextToken.UserID).
		First(&data, form.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrForbidden(err.Error())
		}
		return ErrInternalServerError(err.Error())
	}

	// 只要parentID 不为空 跟 (ids 或 delete) 不为空
	var idsAbove []int64
	if data.ParentID != nil && (len(form.IDs) != 0 || form.Delete != "") {
		idsAbove, err = repository.SearchForParentSeriesIDs(mdb, data.ParentID)
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
	}

	switch form.Delete {
	// 级联
	case "DeleteCASCADE":
		err = mdb.Transaction(func(tx *gorm.DB) (err error) {
			// 上层关联数据
			if len(idsAbove) != 0 {
				err = tx.Model(&entity.Series{}).Where(IN(IDPlain), idsAbove).
					Update(VideosCountPlain, gorm.Expr(Subtract(VideosCountPlain), data.VideosCount)).
					Update(ViewsCountPlain, gorm.Expr(Subtract(ViewsCountPlain), data.ViewsCount)).
					Error
				if err != nil {
					return
				}
			}
			// 直接删除就行
			return tx.Unscoped().Where(Equal(IDPlain), form.ID).Delete(&entity.Series{}).Error
		})
	// 不是级联
	case "Delete":
		var table string
		var ids []int64
		if data.Layer == 0 {
			// 下层查询
			datas := []entity.Video{}
			err = mdb.Select(IDPlain).Where(Equal(SeriesIDPlain), data.ID).Find(&datas).Error
			if err != nil {
				return ErrInternalServerError(err.Error())
			}
			for i := 0; i < len(datas); i++ {
				ids = append(ids, datas[i].ID)
			}
			table = VideoPlain
		} else {
			// 下层查询
			datas := []entity.Series{}
			err = mdb.Select(IDPlain).Where(Equal(SeriesIDPlain), data.ID).Find(&datas).Error
			if err != nil {
				return ErrInternalServerError(err.Error())
			}
			for i := 0; i < len(datas); i++ {
				ids = append(ids, datas[i].ID)
			}
			table = SeriesPlain
		}

		err = mdb.Transaction(func(tx *gorm.DB) (err error) {
			// 需要下层主动Set Null, 对下层的
			switch table {
			case VideoPlain:
				err = tx.Table(table).Where(IN(IDPlain), ids).Update(SeriesIDPlain, gorm.Expr("NULL")).Error
			default:
				err = tx.Table(table).Where(IN(IDPlain), ids).Update(ParentIDPlain, gorm.Expr("NULL")).Error
			}
			if err != nil {
				return err
			}

			// 更新上层关联
			if len(idsAbove) != 0 {
				err = tx.Model(&entity.Series{}).Where(IN(IDPlain), idsAbove).
					Update(VideosCountPlain, gorm.Expr(Subtract(VideosCountPlain), data.VideosCount)).
					Update(ViewsCountPlain, gorm.Expr(Subtract(ViewsCountPlain), data.ViewsCount)).
					Error
				if err != nil {
					return
				}
			}

			// 自我 DeletedAT
			return tx.Delete(&entity.Series{}, data.ID).Error
		})

		if err != nil {
			return ErrInternalServerError(err.Error())
		}
	// 更新部分
	default:
		var idsAppend, idsRemove []int64
		var aDetle repository.UpdateDetle
		if len(form.IDs) != 0 {
			if data.Layer == 0 {
				idsAppend, idsRemove, aDetle, ierr = repository.SecondSearchVideo(mdb, &data, form.IDs)
			} else {
				idsAppend, idsRemove, aDetle, ierr = repository.SecondSearchSeries(mdb, &data, form.IDs)
			}
			if !ierr.IsZero() {
				return
			}
		}
		err := mdb.Transaction(func(tx *gorm.DB) (err error) {
			// 下层关联更新
			if data.Layer == 0 {
				if len(idsRemove) != 0 {
					err = tx.Model(&entity.Video{}).Where(IN(IDPlain), idsRemove).Update(SeriesIDPlain, gorm.Expr("NULL")).Error
					if err != nil {
						return err
					}

				}
				if len(idsAppend) != 0 {
					err = tx.Model(&entity.Video{}).Where(IN(IDPlain), idsAppend).Update(SeriesIDPlain, data.ID).Error
					if err != nil {
						return err
					}
				}
			} else {
				if len(idsRemove) != 0 {
					err = tx.Model(&entity.Series{}).Where(IN(IDPlain), idsRemove).Update(ParentIDPlain, gorm.Expr("NULL")).Error
					if err != nil {
						return err
					}

				}
				if len(idsAppend) != 0 {
					err = tx.Model(&entity.Series{}).Where(IN(IDPlain), idsAppend).Update(ParentIDPlain, data.ID).Error
					if err != nil {
						return err
					}
				}
			}

			// 上层更新,包括自己
			if aDetle.ViewsDetla != 0 || aDetle.VideosCountDetla != 0 {
				err = tx.Model(&entity.Series{}).Where(IN(IDPlain), append(idsAbove, data.ID)).
					Update(ViewsCountPlain, gorm.Expr(Subtract(ViewsCountPlain), -aDetle.ViewsDetla)).
					Update(VideosCountPlain, gorm.Expr(Subtract(VideosCountPlain), -aDetle.VideosCountDetla)).Error
				if err != nil {
					return err
				}
			}

			// 自我字段更新
			if form.Name != "" || form.Description != "" {
				err = tx.Model(&entity.Series{}).Where(Equal(IDPlain), data.ID).
					Updates(&entity.Series{Name: form.Name, Description: form.Description}).Error
				if err != nil {
					return err
				}
			}
			return
		})
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
	}
	c.Set(constant.Result, "ok")
	return
}

func (v VideoService) RandomRecommendVideo(mdb *gorm.DB, buf *videobuf.VideoBufServer) func(c *gin.Context) {
	return func(c *gin.Context) {
		ierr := v.randomRecommendVideo(c, mdb, buf)
		if !ierr.IsZero() {
			base.AddResultIErrAndArgs(c, ierr)
			c.Abort()
			return
		}
	}

}

func (v VideoService) randomRecommendVideo(c *gin.Context, mdb *gorm.DB, buf *videobuf.VideoBufServer) (ierr httperr.IError) {

	contextToken := c.MustGet(constant.ContextToken).(entity.ContextToken)

	if contextToken.UserID != 0 {
		aUser := entity.UserV1{}
		err := mdb.First(&aUser, contextToken.UserID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBadRequest(err.Error())
			}
			return ErrInternalServerError(err.Error())
		}
		if len(aUser.Bytes) == 0 { // 可能空,没有任何偏好
			res, err := buf.GeneralRes()
			if err != nil {
				return ErrInternalServerError(err.Error())
			}
			ginData(c, res)
			return
		}
		bytes := entity.Bytes{}
		err = json.Unmarshal(aUser.Bytes, &bytes)
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
		res, err := buf.GetRes(aUser.ID, &bytes.Video)
		if err != nil {
			return ErrInternalServerError(err.Error())
		}
		ginData(c, res)
		return
	}

	// 如果没有的话
	res, err := buf.GeneralRes()
	if err != nil {
		return ErrInternalServerError(err.Error())
	}
	ginData(c, res)
	return
}
