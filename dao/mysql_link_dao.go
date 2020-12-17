package dao

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"peg.nu/short/model"
)

type MySqlLinkDao struct {
	db *sql.DB
}

func NewMySqlLinkDao(dbInfo DbConnectionInfo) LinkDAO {
	db, err := dbInfo.OpenMySQL()
	if err != nil {
		log.Fatal(err)
	}

	return &MySqlLinkDao{db: db}
}

func (m MySqlLinkDao) Create(link model.Link) bool {
	exists := m.Exists(link.Short)

	_, err := m.db.Exec("insert into link values (?, ?, ?, ?) on duplicate key update `long` = ?, user = ?, user_name = ?", link.Short, link.Long, link.UserId, link.UserName, link.Long, link.UserId, link.UserName)
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
	var long, user, userName string
	err := m.db.QueryRow("select `long`, user, user_name from link where short = ?", short).Scan(&long, &user, &userName)
	if err != nil {
		return model.Link{}, err
	}

	return model.Link{
		Short:    short,
		Long:     long,
		UserId:   user,
		UserName: userName,
	}, nil
}

func (m MySqlLinkDao) GetUserLinks(user string) ([]model.Link, error) {
	rows, err := m.db.Query("select `short`, `long`, `user`, `user_name` from link where `user` = ?", user)
	if err != nil {
		return []model.Link{}, err
	}
	defer rows.Close()

	var links []model.Link
	for rows.Next() {
		var short, long, userId, userName string

		if err := rows.Scan(&short, &long, &userId, &userName); err != nil {
			return []model.Link{}, err
		}

		links = append(links, model.Link{
			Short:    short,
			Long:     long,
			UserId:   userId,
			UserName: userName,
		})
	}

	return links, nil
}

func (m MySqlLinkDao) Close() {
	err := m.db.Close()
	if err != nil {
		log.Fatal(err)
	}
}
