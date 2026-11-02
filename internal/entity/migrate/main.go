package main

import (
	"all-in-one/config"
	"all-in-one/internal/entity"
	"all-in-one/pkg/mysqlx"
	"fmt"
)

var (
	mdb = mysqlx.NewDB(config.Conf.Mysql)
)

func main() {
	err := mdb.AutoMigrate(
		// &entity.CollectedVideo{},
		//  &entity.LoginLocation{},
		// &entity.Comment{},
		// &entity.UserFollow{},
		// &entity.View{},
		// &entity.Video{},
		// &entity.Series{},
		&entity.Active{},
	)
	if err != nil {
		fmt.Println(err)
	}
}
