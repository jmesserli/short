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
	"peg.nu/short/dao"
	"peg.nu/short/shortener"
	"peg.nu/short/unsplash"
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

	u := unsplash.New(
		os.Getenv("SHORT_UNSPLASH_ACCESSKEY"),
		os.Getenv("SHORT_DB_HOST"),
		os.Getenv("SHORT_DB_DATABASE"),
		os.Getenv("SHORT_DB_USER"),
		os.Getenv("SHORT_DB_PASS"))

	r := mux.NewRouter()

	r.HandleFunc("/api/link", s.CreateLink).Methods("POST")
	r.HandleFunc("/api/link/{link}", s.DeleteLink).Methods("DELETE")
	r.HandleFunc("/api/unsplash/image", u.GetImage).Methods("GET")

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
		var requiredRole string
		if strings.HasPrefix(r.URL.Path, "/api/link") {
			method := r.Method

			if method == http.MethodPost {
				requiredRole = "PegNu-Short.CREATE"
			} else if method == http.MethodDelete {
				requiredRole = "PegNu-Short.DELETE"
			}

			needsLogin = true
		} else if strings.HasPrefix(r.URL.Path, "/api/unsplash") {
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
		token := r.Context().Value("user").(*jwt.Token)
		rw.Header().Add("X-Short-Sub", token.Claims.(jwt.MapClaims)["sub"].(string))
		rw.Header().Add("X-Short-User", token.Claims.(jwt.MapClaims)["preferred_username"].(string))

		// short circuit if no role is required
		if len(requiredRole) == 0 {
			next(rw, r)
			return
		}

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
