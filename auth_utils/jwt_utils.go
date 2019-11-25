package auth_utils

import (
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
)

type MaybeUser struct {
	IsAuthenticated bool
	User            *User
}

type User struct {
	Id        string
	GivenName string
	FirstName string
	LastName  string
	Username  string
	Email     string
	Scopes    []string
	Roles     map[string][]string
}

func (u User) HasRole(client string, role ShortRole) bool {
	return u.HasRoles(client, []ShortRole{role})
}

func (u User) HasRoles(client string, roles []ShortRole) bool {
	for _, neededRole := range roles {
		found := false

		for _, role := range u.Roles[client] {
			if role == string(neededRole) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func GetUser(r *http.Request) *MaybeUser {
	ctxUser := r.Context().Value("user")

	if ctxUser == nil {
		return &MaybeUser{IsAuthenticated: false}
	}

	return &MaybeUser{
		IsAuthenticated: true,
		User:            ParseToken(ctxUser.(*jwt.Token)),
	}
}

func ParseToken(token *jwt.Token) *User {
	claims := token.Claims.(jwt.MapClaims)

	user := User{
		Id:        claims["sub"].(string),
		GivenName: claims["name"].(string),
		FirstName: claims["given_name"].(string),
		LastName:  claims["family_name"].(string),
		Username:  claims["preferred_username"].(string),
		Email:     claims["email"].(string),
	}

	rolesMap := map[string][]string{}
	resourceAccess := claims["resource_access"].(map[string]interface{})

	for resource, resMap := range resourceAccess {
		roleMap := resMap.(map[string]interface{})
		roles := roleMap["roles"].(map[string]string)
		strRoles := make([]string, len(roles))

		for _, role := range roles {
			strRoles = append(strRoles, role)
		}

		rolesMap[resource] = strRoles
	}
	user.Roles = rolesMap
	user.Scopes = strings.Split(claims["scope"].(string), " ")

	return &user
}
