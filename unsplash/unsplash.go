package unsplash

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	us "github.com/hbagdi/go-unsplash/unsplash"
	"log"
	"net/http"
	"peg.nu/short/dao"
	"peg.nu/short/global"
	"peg.nu/short/model"
	"time"
)

type authenticatingTransport struct {
	accessKey string
}

func (at authenticatingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %v", at.accessKey))

	return http.DefaultTransport.RoundTrip(req)
}

type Unsplash struct {
	client *us.Unsplash
	dao    dao.UnsplashDAO
}

func New(accessKey string, dao dao.UnsplashDAO) Unsplash {
	hc := http.Client{Transport: authenticatingTransport{accessKey: accessKey}}

	return Unsplash{
		client: us.New(&hc),
		dao:    dao,
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
	err, _ := u.dao.Update(model.Image{
		ImageUrl:             "",
		PhotographerName:     "",
		PhotographerUsername: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	body, _ := json.Marshal(map[string]string{
		"status": "ok",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, err = w.Write(body)
	if err != nil {
		panic(err)
	}
}

func (u Unsplash) getImage() model.Image {
	err, dbImg := u.dao.Get()
	if err != nil {
		log.Fatal(err)
	}

	if dbImg.ImageUrl != "" && dbImg.Updated.Add(global.Expiration).After(time.Now()) {
		return *dbImg
	}

	image := u.queryRandomImage()
	err, _ = u.dao.Update(image)
	if err != nil {
		log.Fatal(err)
	}

	return image
}

func (u Unsplash) queryRandomImage() model.Image {
	photos, _, err := u.client.Photos.Random(&us.RandomPhotoOpt{
		Orientation:   us.Landscape,
		Count:         1,
		CollectionIDs: []int{573009},
	})
	if err != nil {
		log.Fatal(err)
	}

	photo := (*photos)[0]
	return model.Image{
		ImageUrl:             photo.Urls.Full.String(),
		PhotographerName:     *photo.Photographer.Name,
		PhotographerUsername: *photo.Photographer.Username,
		Updated:              time.Now(),
		ExpirationDuration:   global.Expiration,
	}
}
