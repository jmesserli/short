package shortener

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"peg.nu/short/auth_utils"
	"peg.nu/short/dao"
	"peg.nu/short/model"
	"regexp"
	"strings"
	"time"
)

func returnJson(w http.ResponseWriter, code int, data interface{}) {
	strVal, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(strVal)
	if err != nil {
		panic(err)
	}
}

func returnError(w http.ResponseWriter, err error, code int) {
	m := map[string]string{
		"status":  "error",
		"message": err.Error(),
	}

	returnJson(w, code, m)
}

type Shortener struct {
	Dao    dao.LinkDAO
	Random rand.Rand
}

func NewShortener(dao dao.LinkDAO) *Shortener {
	return &Shortener{
		Dao:    dao,
		Random: *rand.New(rand.NewSource(time.Now().Unix())),
	}
}

var generateChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.")

func (s Shortener) generateUniqueLink() string {
	length := 6

	makeRandom := func() string {
		runes := make([]rune, 0, length)
		totalChars := len(generateChars)

		for i := 0; i < length; i++ {
			runes = append(runes, generateChars[s.Random.Intn(totalChars)])
		}
		return string(runes)
	}

	var randomLink string
	for len(randomLink) == 0 || s.Dao.Exists(randomLink) {
		randomLink = makeRandom()
	}

	return randomLink
}

var shortRegex = regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_\\-]{0,62}[a-zA-Z0-9]$")

func (s Shortener) LinkExists(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	exists := s.Dao.Exists(vars["link"])
	if !exists {
		returnJson(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
			"exists": false,
		})
		return
	}

	link, err := s.Dao.Get(vars["link"])
	if err != nil {
		returnError(w, err, 500)
		return
	}

	returnJson(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"exists":    true,
		"url":       link.Long,
		"user":      link.UserId,
		"user_name": link.UserName,
	})
}

func (s Shortener) CreateLink(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		returnError(w, err, 500)
		return
	}

	link := model.Link{}
	err = json.Unmarshal(bodyBytes, &link)
	if err != nil {
		returnError(w, err, http.StatusBadRequest)
		return
	}

	user := auth_utils.GetUser(r).User
	link.UserId = user.Id
	link.UserName = user.Username

	if len(link.Short) > 0 && s.Dao.Exists(link.Short) && !user.HasRole("short", auth_utils.RoleOverwrite) {
		savedLink, _ := s.Dao.Get(link.Short)
		if savedLink.UserId != user.Id {
			returnError(w, fmt.Errorf("you don't have permisson to overwrite links"), http.StatusForbidden)
			return
		}
	}

	if len(link.Short) == 0 {
		link.Short = s.generateUniqueLink()
	}

	if !shortRegex.Match([]byte(link.Short)) {
		returnError(w, fmt.Errorf("invalid short link format"), http.StatusBadRequest)
		return
	}

	_, err = url.Parse(link.Long)
	if err != nil || len(strings.TrimSpace(link.Long)) == 0 {
		returnError(w, err, http.StatusBadRequest)
		return
	}

	existed := s.Dao.Create(link)
	returnJson(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"existed": existed,
		"link":    link,
	})
}

func (s Shortener) DeleteLink(w http.ResponseWriter, r *http.Request) {
	short := mux.Vars(r)["link"]

	if !s.Dao.Exists(short) {
		returnJson(w, http.StatusNotFound, map[string]string{
			"status": "not_found",
		})
		return
	}

	user := auth_utils.GetUser(r).User
	userCanDeleteGlobal := user.HasRole("short", auth_utils.RoleDelete)

	link, err := s.Dao.Get(short)
	if err != nil {
		returnError(w, fmt.Errorf("failed retrieving link"), http.StatusInternalServerError)
		return
	}

	if link.UserId != user.Id && !userCanDeleteGlobal {
		returnError(w, fmt.Errorf("you are not authorized to delete other user's links"), http.StatusForbidden)
		return
	}

	s.Dao.Delete(short)

	returnJson(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s Shortener) RedirectShort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	short := vars["path"]

	link, err := s.Dao.Get(short)
	if err != nil {
		returnError(w, err, 404)
		return
	}

	w.Header().Set("Location", link.Long)
	w.Header().Set("x-short-link", link.Short)
	w.Header().Set("x-short-user", link.UserId)
	w.WriteHeader(http.StatusFound)
}

func (s Shortener) UserLinks(w http.ResponseWriter, r *http.Request) {
	user := auth_utils.GetUser(r).User

	links, err := s.Dao.GetUserLinks(user.Id)
	if err != nil {
		returnError(w, fmt.Errorf("failed to retrieve user links"), http.StatusInternalServerError)
		return
	}

	returnJson(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"links":  links,
	})
}
