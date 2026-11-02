package repository

import (
	"all-in-one/internal/base/constant"
	"all-in-one/internal/entity"
	httperr "all-in-one/pkg/ekit/http_err"
	"sort"

	"gorm.io/gorm"
)

const (
	updateRecordNotFound             = "updateRecordNotFound"
	attempToUpdateOtherSeries        = "attempToUpdateOtherSeries"
	attempToUpdateOtherLayer         = "attempToUpdateOtherLayer"
	layerNotMatch                    = "layerNotMatch"
	attempToUpdateOtherVideo         = "attempToUpdateOtherVideo"
	attempToSeiresVideoThatHasSeries = "attempToSeiresVideoThatHasSeries"
	mdbVideoRecordNotFound           = "mdbVideoRecordNotFound"
)

// 二次查询对应的video
// 获得对应有效ids
// idsAppend, idsRemove, ierr
func SecondSearchVideo(mdb *gorm.DB, data *entity.Series, targetIDs []int64) ([]int64, []int64, UpdateDetle, httperr.IError) {
	datas := []entity.Video{}
	err := mdb.
		Select(ViewsCountPlain, UserIDPlain, SeriesIDPlain, IDPlain).
		Where("SeriesID = ?", data.ID).
		Or("ID IN ?", targetIDs).
		Order("ID").
		Find(&datas).Error
	if err != nil {
		return []int64{}, []int64{}, UpdateDetle{}, httperr.ErrInternalServerError(err.Error())
	}
	// 比对一下
	idsAppend := []int64{}
	idsRemove := []int64{}
	aDetle := UpdateDetle{}
	sort.Slice(targetIDs, func(i, j int) bool { return targetIDs[i] < targetIDs[j] })
	l1, l2 := len(datas), len(targetIDs)
	var i, j int
	for i < l1 && j < l2 {
		// 更新别人的
		if datas[i].UserID != data.UserID {
			return []int64{}, []int64{}, aDetle, httperr.ErrForbidden(attempToUpdateOtherVideo)
		}
		// 非法添加个人其他的video
		if datas[i].SeriesID != nil && *datas[i].SeriesID != data.ID {
			return []int64{}, []int64{}, aDetle, httperr.ErrForbidden(attempToSeiresVideoThatHasSeries)
		}
		// 如果form的小了,说明mdb没查到,啥概念?无用数据,或者非法吧,
		if datas[i].ID > targetIDs[j] {
			return []int64{}, []int64{}, aDetle, httperr.ErrBadRequest(mdbVideoRecordNotFound)
		}
		// 只需要命中id就行
		if datas[i].ID == targetIDs[j] {
			if datas[i].SeriesID == nil {
				aDetle.ViewsDetla += datas[i].ViewsCount
				idsAppend = append(idsAppend, datas[i].ID)
			}
			i++
			j++
			continue
		}
		// 小于的部分,db查到,但不再需要,删除

		aDetle.ViewsDetla -= datas[i].ViewsCount
		idsRemove = append(idsRemove, datas[i].ID)
		i++
	}
	// 必定非法,部分没找到
	if j != l2 {
		return []int64{}, []int64{}, aDetle, httperr.ErrBadRequest(mdbVideoRecordNotFound)
	}
	if i != l1 {
		// 这部分是mdb有但没发过来,这完全保证是user自己的了
		for i1 := i; i < l1; i++ {
			aDetle.ViewsDetla -= datas[i1].ViewsCount
			idsRemove = append(idsRemove, datas[i1].ID)
		}
	}
	aDetle.VideosCountDetla = int64(len(idsAppend) - len(idsRemove))
	return idsAppend, idsRemove, aDetle, httperr.IError{}
}

// 二次查询对应的series
// 获得对应有效ids
// idsAppend, idsRemove, ierr
func SecondSearchSeries(mdb *gorm.DB, data *entity.Series, targetIDs []int64) ([]int64, []int64, UpdateDetle, httperr.IError) {
	datas := []entity.Series{}
	err := mdb.
		Select(ViewsCountPlain, UserIDPlain, ParentIDPlain, IDPlain, VideosCountPlain, LayerPlain).
		Where("ParentID = ?", data.ID).
		Or("ID IN ?", targetIDs).
		Order(IDPlain).
		Find(&datas).Error
	if err != nil {
		return []int64{}, []int64{}, UpdateDetle{}, httperr.ErrInternalServerError(err.Error())
	}
	// 比对一下
	idsAppend := []int64{}
	idsRemove := []int64{}
	aDetle := UpdateDetle{}
	sort.Slice(targetIDs, func(i, j int) bool { return targetIDs[i] < targetIDs[j] })
	l1, l2 := len(datas), len(targetIDs)
	var i, j int
	for i < l1 && j < l2 {
		// 不同的layer
		if data.Layer != datas[i].Layer+1 {
			return []int64{}, []int64{}, aDetle, httperr.ErrForbidden(attempToUpdateOtherLayer)
		}
		// 更新别人的
		if datas[i].UserID != data.UserID {
			return []int64{}, []int64{}, aDetle, httperr.ErrForbidden(attempToUpdateOtherVideo)
		}
		// 非法添加个人其他的video
		if datas[i].ParentID != nil && *datas[i].ParentID != data.ID {
			return []int64{}, []int64{}, aDetle, httperr.ErrForbidden(attempToSeiresVideoThatHasSeries)
		}
		// 如果form的小了,说明mdb没查到,啥概念?无用数据,或者非法吧,
		if datas[i].ID > targetIDs[j] {
			return []int64{}, []int64{}, aDetle, httperr.ErrBadRequest(mdbVideoRecordNotFound)
		}
		// 只需要命中id就行
		// 只需要命中id就行
		if datas[i].ID == targetIDs[j] {
			if datas[i].ParentID == nil {
				aDetle.ViewsDetla += datas[i].ViewsCount
				aDetle.VideosCountDetla += datas[i].VideosCount
				idsAppend = append(idsAppend, datas[i].ID)
			}
			i++
			j++
			continue
		}
		// 小于的部分,db查到,但不再需要,删除
		aDetle.ViewsDetla -= datas[i].ViewsCount
		aDetle.VideosCountDetla -= datas[i].VideosCount
		idsRemove = append(idsRemove, datas[i].ID)
		i++
	}
	// 必定非法,部分没找到
	if j != l2 {
		return []int64{}, []int64{}, aDetle, httperr.ErrBadRequest(mdbVideoRecordNotFound)
	}
	if i != l1 {
		// 这部分是mdb有但没发过来,这完全保证是user自己的了
		for i1 := i; i < l1; i++ {
			aDetle.ViewsDetla -= datas[i1].ViewsCount
			aDetle.VideosCountDetla -= datas[i1].VideosCount
			idsRemove = append(idsRemove, datas[i1].ID)
		}
	}
	return idsAppend, idsRemove, aDetle, httperr.IError{}
}

// 可以传入video的直接seriesID,当然也就直接携带直接的seriesID一起返回
// 返回 gorm.err
func SearchForParentSeriesIDs(mdb *gorm.DB, seriesId *int64) ([]int64, error) {
	iter := 0
	parentID := seriesId
	ids := make([]int64, 0)
	var err error
	for parentID != nil && iter < constant.DBSearchMaxIter {
		cur := entity.Series{}
		err = mdb.First(&cur, parentID).Error
		if err != nil {
			break
		}
		ids = append(ids, cur.ID)
		if cur.ParentID == nil {
			break
		}
		parentID = cur.ParentID
		iter++
	}
	return ids, err
}

type UpdateDetle struct {
	ViewsDetla       int64
	VideosCountDetla int64
}
