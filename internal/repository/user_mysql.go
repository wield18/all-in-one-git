package repository

import (
	"all-in-one/internal/base"
	"all-in-one/internal/entity"
	"all-in-one/pkg/consumer"
	httperr "all-in-one/pkg/ekit/http_err"
	"context"
	"log/slog"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	passwordOrUserWrong = "passwordOrUserWrong"
)

func UserLocaionCreateBatch(mysqlDB *gorm.DB) func(ctx context.Context, datas []*consumer.ConsumeData[entity.LoginLocation]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[entity.LoginLocation]) {
		dats := []*entity.LoginLocation{}
		for _, data := range datas {
			if !data.Err.IsZero() {
				continue
			}
			dats = append(dats, data.Data.EncodeDB())
		}
		if err := mysqlDB.CreateInBatches(dats, len(dats)).Error; err != nil {
			for _, data := range datas {
				if !data.Err.IsZero() {
					data.WithErr(httperr.ErrInternalServerError(err.Error()))
				}
			}
		}
	}
}

// 内部进行 500 slog
// 只有正确与否
func CheckPassword(mdb *gorm.DB, id int64, password string) bool {
	aUser := entity.UserV1{}
	err := mdb.Select(aUser.CheckPasswordSelect()).First(&aUser, id).Error
	if err != nil || !base.CheckCryptoCode(aUser.PasswordCrypto, password) {
		return false
	}
	return true
}

func UserViewVideos(mysqlDB *gorm.DB) func(ctx context.Context, datas []*consumer.ConsumeData[entity.View]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[entity.View]) {
		for _, data := range datas {
			if !data.Err.IsZero() {
				continue
			}
			err := mysqlDB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "VideoID"}, {Name: "UserID"}},
				DoUpdates: clause.AssignmentColumns([]string{"LastViewTime"}),
			}).Create(&data.Data).Error
			if err != nil {
				data.WithErr(httperr.ErrInternalServerError(err.Error()))
			}
		}

	}
}

func VideoCommentCount(mysqlDB *gorm.DB) func(ctx context.Context, datas []*consumer.ConsumeData[entity.VideoDetle]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[entity.VideoDetle]) {
		dats := make(map[int64]int64)
		for _, v := range datas {
			if _, ok := dats[v.Data.ID]; !ok {
				dats[v.Data.ID] = v.Data.Detle
				continue
			}
			dats[v.Data.ID] += v.Data.Detle
		}
		var err error
		for videoID, CommentCount := range dats {
			err = mysqlDB.Model(entity.Video{}).
				Where(Equal(IDPlain), videoID).Update(CommentsCountPlain, gorm.Expr(Subtract(CommentsCountPlain), -CommentCount)).Error
		}
		if err != nil {
			for _, v := range datas {
				v.WithErr(httperr.ErrInternalServerError(err.Error()))
			}
		}

	}
}

// 这只是创建当前天的 active, 重复更新lastActiveTime
func UserActive(mysqlDB *gorm.DB) func(ctx context.Context, datas []*consumer.ConsumeData[*entity.Active]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[*entity.Active]) {
		for _, data := range datas {
			if !data.Err.IsZero() {
				continue
			}
			err := mysqlDB.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "UserID"},
					{Name: "Days"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"LastActiveTime"}),
			}).Create(data.Data).Error
			if err != nil {
				data.WithErr(httperr.ErrInternalServerError(err.Error()))
			}
		}

	}
}

func IncrVideoViewsCount(mysqlDB *gorm.DB, aMap map[int64]int64) {
	// 构建 CASE WHEN 片段
	var caseSQL strings.Builder
	args := make([]interface{}, 0, 2*len(aMap))
	ids := make([]int64, 0, len(aMap))
	for k, v := range aMap {
		caseSQL.WriteString("WHEN ? THEN ViewsCount + ? ")
		args = append(args, k, v)
		ids = append(ids, k)
	}
	args = append(args, ids)
	// 拼接最终 SQL
	sql := "UPDATE Video SET ViewsCount = CASE ID " + caseSQL.String() + "END WHERE ID IN (?)"

	err := mysqlDB.Exec(sql, args...).Error
	if err != nil {
		slog.Error(err.Error())
	}
}

func IncrActiveCount(mysqlDB *gorm.DB, aMap map[int64]int64) {
	// 构建 CASE WHEN 片段
	var caseSQL strings.Builder
	args := make([]interface{}, 0, 2*len(aMap)+1)
	ids := make([]int64, 0, len(aMap))
	for k, v := range aMap {
		caseSQL.WriteString("WHEN ? THEN DaysActiveCount + ? ")
		args = append(args, k, v)
		ids = append(ids, k)
	}
	args = append(args, ids)
	// 拼接最终 SQL
	sql := "UPDATE Active SET DaysActiveCount = CASE ID " + caseSQL.String() + "END WHERE ID IN (?)"

	err := mysqlDB.Exec(sql, args...).Error
	if err != nil {
		slog.Error(err.Error())
	}
}
