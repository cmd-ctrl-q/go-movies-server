package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cmd-ctrl-q/go-movies-server/models"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
)

var validUser = models.User{
	ID:       10,
	Email:    "me@here.com",
	Password: createHash("password"),
}

type Credentials struct {
	Username string `json:"email"`
	Password string `json:"password"`
}

func (app *application) Signin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.logger.Println("unauthorized user at signin")
		app.errorJSON(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	// get username and password, query db and check if there is a valid user.
	// and check the hashed password against the hash in the db.
	// mock this process:
	hashedPassword := validUser.Password

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		app.logger.Println("unauthorized user at signin")
		app.errorJSON(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	// create and add claims
	// user is valid
	var claims jwt.Claims
	claims.Subject = fmt.Sprintf("%d", validUser.ID)                    // claim subject
	claims.Issued = jwt.NewNumericTime(time.Now())                      // when was claim issued
	claims.NotBefore = jwt.NewNumericTime(time.Now())                   // not valid before now
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour)) // when does it expire (24 hours)
	claims.Issuer = "mydomain.com"
	claims.Audiences = []string{"mydomain.com"} // who can see this

	// create token
	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	if err != nil {
		app.logger.Println("error signing")
		app.errorJSON(w, http.StatusUnauthorized, errors.New("error signing"))
		return
	}

	// write to client
	app.writeJSON(w, http.StatusOK, jwtBytes, "response")
}

func createHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash)
}
