package repository

import (
	"all-in-one/internal/base/constant"
	httperr "all-in-one/pkg/ekit/http_err"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func IsErrDuplicated(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return false
}

func MdbErrNotFound(err error) httperr.IError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return httperr.ErrNotFound(err.Error())
	}
	return httperr.ErrInternalServerError(err.Error())
}

func Paging(mdb *gorm.DB, limitTime int64, page int, orderBy string, desc bool) *gorm.DB {
	// 有时间限制就写上
	if limitTime != 0 {
		mdb = mdb.Where(Greater(CreatedAtPlain), time.Now().Add(-(time.Duration(limitTime) * time.Second)))
	}
	mdb = mdb.Offset(page * constant.PageSize).Limit(constant.PageSize)
	if orderBy != "" {
		mdb = mdb.Order(clause.OrderByColumn{Column: clause.Column{Name: orderBy}, Desc: desc})
	}
	return mdb
}
