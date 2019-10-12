package dao

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"peg.nu/short/model"
)

type MySqlLinkDao struct {
	db *sql.DB
}

func NewMySqlLinkDao(host, database, user, password string) LinkDAO {
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v)/%v", user, password, host, database))
	if err != nil {
		log.Fatal(err)
	}

	return &MySqlLinkDao{db: db,}
}

func (m MySqlLinkDao) Create(link model.Link) bool {
	exists := m.Exists(link.Short)

	_, err := m.db.Exec("insert into link values (?, ?) on duplicate key update `long` = ?", link.Short, link.Long, link.Long)
	if err != nil {
		log.Fatal(err)
	}

	return exists
}

func (m MySqlLinkDao) Exists(short string) bool {
	var count int
	err := m.db.QueryRow("select count(*) from link where short = ?", short).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	return count > 0
}

func (m MySqlLinkDao) Delete(short string) bool {
	result, err := m.db.Exec("delete from link where short = ?", short)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	return rowsAffected > 0
}

func (m MySqlLinkDao) Get(short string) (model.Link, error) {
	var long string
	err := m.db.QueryRow("select `long` from link where short = ?", short).Scan(&long)
	if err != nil {
		return model.Link{}, err
	}

	return model.Link{
		Short: short,
		Long:  long,
	}, nil
}

func (m MySqlLinkDao) Close() {
	err := m.db.Close()
	if err != nil {
		log.Fatal(err)
	}
}
