package main

import (
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"net/http"
	"os"
	"peg.nu/short/dao"
	"peg.nu/short/shortener"
	"time"
)

func main() {
	s := shortener.NewShortener(
		dao.NewMySqlLinkDao(
			os.Getenv("SHORT_DB_HOST"),
			os.Getenv("SHORT_DB_DATABASE"),
			os.Getenv("SHORT_DB_USER"),
			os.Getenv("SHORT_DB_PASS")))

	r := mux.NewRouter()

	r.HandleFunc("/api/link", s.CreateLink).Methods("POST")
	r.HandleFunc("/api/link/{link}", s.DeleteLink).Methods("DELETE")

	r.HandleFunc("/{path}", s.RedirectShort).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
