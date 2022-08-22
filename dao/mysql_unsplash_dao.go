package dao

import (
	"database/sql"
	"log"
	"peg.nu/short/global"
	"peg.nu/short/model"
)

type MysqlUnsplashDao struct {
	db *sql.DB
}

func NewMysqlUnsplashDao(dbInfo DbConnectionInfo) UnsplashDAO {
	db, err := dbInfo.OpenMySQL()
	if err != nil {
		log.Fatal(err)
	}

	return &MysqlUnsplashDao{db: db}
}

func (m MysqlUnsplashDao) Get() (error, *model.Image) {
	img := model.Image{}

	err := m.db.QueryRow("select url, photographer_name, photographer_profile, updated from unsplash_image where id = ?", 1).Scan(&img.ImageUrl, &img.PhotographerName, &img.PhotographerUsername, &img.Updated)
	if err != nil {
		return err, nil
	}

	img.ExpirationDuration = global.Expiration

	return nil, &img
}

func (m MysqlUnsplashDao) Update(img model.Image) (error, bool) {
	res, err := m.db.Exec("update unsplash_image set url = ?, photographer_name = ?, photographer_profile = ?, updated = NOW() where id = ?", img.ImageUrl, img.PhotographerName, img.PhotographerUsername, 1)
	if err != nil {
		return err, false
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err, false
	}

	return nil, rows > 0
}
