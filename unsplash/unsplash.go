package unsplash

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	us "github.com/hbagdi/go-unsplash/unsplash"
	"log"
	"net/http"
	"time"
)

type authenticatingTransport struct {
	accessKey string
}

func (at authenticatingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %v", at.accessKey))

	return http.DefaultTransport.RoundTrip(req)
}

type Image struct {
	ImageUrl             string    `json:"image_url"`
	PhotographerName     string    `json:"photographer_name"`
	PhotographerUsername string    `json:"photographer_username"`
	Updated              time.Time `json:"-"`
}

type Unsplash struct {
	client *us.Unsplash
	db     *sql.DB
}

func New(accessKey, host, database, user, password string) Unsplash {
	hc := http.Client{Transport: authenticatingTransport{accessKey: accessKey}}

	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v)/%v?parseTime=true", user, password, host, database))
	if err != nil {
		log.Fatal(err)
	}

	return Unsplash{
		client: us.New(&hc),
		db:     db,
	}
}

func (u Unsplash) GetImage(w http.ResponseWriter, r *http.Request) {
	image := u.getImage()

	strVal, err := json.Marshal(image)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, err = w.Write(strVal)
	if err != nil {
		panic(err)
	}
}

func (u Unsplash) Clear(w http.ResponseWriter, r *http.Request) {
	u.updateDbImage(Image{
		ImageUrl:             "",
		PhotographerName:     "",
		PhotographerUsername: "",
	})

	body, _ := json.Marshal(map[string]string{
		"status": "ok",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, err := w.Write(body)
	if err != nil {
		panic(err)
	}
}

func (u Unsplash) getImage() Image {
	dbImg := u.getDbImage()

	if dbImg.ImageUrl != "" && dbImg.Updated.Add(12*time.Hour).After(time.Now()) {
		return dbImg
	}

	image := u.queryRandomImage()
	u.updateDbImage(image)

	return image
}

func (u Unsplash) queryRandomImage() Image {
	photos, _, err := u.client.Photos.Random(&us.RandomPhotoOpt{
		Orientation:   us.Landscape,
		Count:         1,
		CollectionIDs: []int{573009},
	})
	if err != nil {
		log.Fatal(err)
	}

	photo := (*photos)[0]
	return Image{
		ImageUrl:             photo.Urls.Full.String(),
		PhotographerName:     *photo.Photographer.Name,
		PhotographerUsername: *photo.Photographer.Username,
		Updated:              time.Now(),
	}
}

func (u Unsplash) getDbImage() Image {
	img := Image{}

	err := u.db.QueryRow("select url, photographer_name, photographer_profile, updated from unsplash_image where id = ?", 1).Scan(&img.ImageUrl, &img.PhotographerName, &img.PhotographerUsername, &img.Updated)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

func (u Unsplash) updateDbImage(img Image) bool {
	res, err := u.db.Exec("update unsplash_image set url = ?, photographer_name = ?, photographer_profile = ?, updated = NOW() where id = ?", img.ImageUrl, img.PhotographerName, img.PhotographerUsername, 1)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	return rows > 0
}
