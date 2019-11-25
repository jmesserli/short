package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
	"peg.nu/short/auth_utils"
	"peg.nu/short/dao"
	"peg.nu/short/shortener"
	"peg.nu/short/unsplash"
	"strings"
	"time"
)

func main() {
	dbInfo := dao.DbConnectionInfo{
		Host:     os.Getenv("SHORT_DB_HOST"),
		Database: os.Getenv("SHORT_DB_DATABASE"),
		User:     os.Getenv("SHORT_DB_USER"),
		Password: os.Getenv("SHORT_DB_PASS"),
	}

	s := shortener.NewShortener(dao.NewMySqlLinkDao(dbInfo))

	u := unsplash.New(os.Getenv("SHORT_UNSPLASH_ACCESSKEY"), dbInfo)

	r := mux.NewRouter()

	r.HandleFunc("/api/link", s.CreateLink).Methods("POST")
	r.HandleFunc("/api/link/{link}", s.DeleteLink).Methods("DELETE")

	r.HandleFunc("/api/unsplash/image", u.GetImage).Methods("GET")
	r.HandleFunc("/api/unsplash/clear", u.Clear).Methods("GET")

	r.HandleFunc("/{path}", s.RedirectShort).Methods("GET")
	r.HandleFunc("/", http.RedirectHandler(os.Getenv("SHORT_UI_URL"), http.StatusFound).ServeHTTP).Methods("GET")

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
		var needsLogin bool
		var requiredRoles []auth_utils.ShortRole
		if strings.HasPrefix(r.URL.Path, "/api/link") {
			method := r.Method

			if method == http.MethodPost {
				requiredRoles = append(requiredRoles, auth_utils.RoleCreate)
			} else if method == http.MethodDelete {
				requiredRoles = append(requiredRoles, auth_utils.RoleDelete)
			}

			needsLogin = true
		} else if strings.HasPrefix(r.URL.Path, "/api/unsplash") {
			if strings.HasPrefix(r.URL.Path, "/api/unsplash/clear") {
				requiredRoles = append(requiredRoles, auth_utils.RoleClearBackground)
			}

			needsLogin = true
		}

		// short circuit if no login is required
		if !needsLogin {
			next(rw, r)
			return
		}

		if r.Context().Value("user") == nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		user := auth_utils.GetUser(r).User
		rw.Header().Add("X-Short-Sub", user.Id)
		rw.Header().Add("X-Short-User", user.Username)

		// short circuit if no role is required
		if len(requiredRoles) == 0 {
			next(rw, r)
			return
		}

		if !user.HasRoles("short", requiredRoles) {
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		next(rw, r)
	}))
	n.UseHandler(r)

	corsHeaders := handlers.AllowedHeaders([]string{"Authorization", "Content-Type"})
	corsOrigins := handlers.AllowedOrigins([]string{"*"})
	corsMethods := handlers.AllowedMethods([]string{"OPTIONS", "GET", "POST", "DELETE"})

	srv := &http.Server{
		Handler:      handlers.CORS(corsHeaders, corsOrigins, corsMethods)(n),
		Addr:         os.Getenv("SHORT_LISTEN_ADDR"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server")
	log.Fatal(srv.ListenAndServe())
}
