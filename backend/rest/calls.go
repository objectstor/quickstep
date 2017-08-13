package qrouter

/*
all rest calls
*/
import (
	"encoding/json"
	"fmt"
	"net/http"
	"quickstep/backend/store"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type QuickStepUserClaims struct {
	Otp string `json:"otp"`
	jwt.StandardClaims
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Org      string `json: "org"` // additional org string ex. shop.com, research_group
}

func doLogin(s *qdb.QSession) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if s == nil {
			JsonError(w, "database not connected ", http.StatusInternalServerError)
			return
		}
		session := s.New()
		defer session.Close()
		var user User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&user)
		if err != nil {
			JsonError(w, "Auth error", http.StatusForbidden)
			return
		}
		if len(user.Name) == 0 || len(user.Password) == 0 {
			JsonError(w, "Auth error", http.StatusForbidden)
			return
		}

		// Create the Claims
		claims := QuickStepUserClaims{
			"OTPPASSWORD", // this chould be changed to otp and verified just idea
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
				Issuer:    "app_name", // this should be taken from config file I think
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(s.SigningKey)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "{\"token\": \"%s\"}", tokenString)
	}
}
