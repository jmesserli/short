package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
	"peg.nu/short/dao"
	"peg.nu/short/shortener"
	"strings"
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

	block, _ := pem.Decode([]byte(fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%v\n-----END PUBLIC KEY-----", os.Getenv("SHORT_JWT_PUBKEY"))))
	if block == nil {
		panic("failed to parse pem block containing public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// should return *rsa.PublicKey for RS256
			return pub, nil
		},
		SigningMethod:       jwt.SigningMethodRS256,
		CredentialsOptional: true,
	})

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.Use(negroni.HandlerFunc(jwtMiddleware.HandlerWithNext))
	n.Use(negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		var requiredRole string
		if strings.HasPrefix(r.URL.Path, "/api/link") {
			method := r.Method

			if method == http.MethodPost {
				requiredRole = "PegNu-Short.CREATE"
			} else if method == http.MethodDelete {
				requiredRole = "PegNu-Short.DELETE"
			}
		}

		// short circuit if no authentication is required
		if len(requiredRole) == 0 {
			next(rw, r)
			return
		}

		if r.Context().Value("user") == nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := r.Context().Value("user").(*jwt.Token)
		resAccess := token.Claims.(jwt.MapClaims)["resource_access"].(map[string]interface{})
		roles := resAccess["short"].(map[string]interface{})["roles"].([]interface{})

		roleFound := false
		for _, role := range roles {
			if role.(string) == requiredRole {
				roleFound = true
				break
			}
		}

		if !roleFound {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		next(rw, r)
	}))
	n.UseHandler(r)

	srv := &http.Server{
		Handler:      n,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
