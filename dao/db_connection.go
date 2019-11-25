package dao

import (
	"database/sql"
	"fmt"
)

type DbConnectionInfo struct {
	Host     string
	Database string
	User     string
	Password string
}

func (dci DbConnectionInfo) OpenMySQL() (*sql.DB, error) {
	return sql.Open(
		"mysql",
		fmt.Sprintf("%v:%v@tcp(%v)/%v?parseTime=true", dci.User, dci.Password, dci.Host, dci.Database))
}
