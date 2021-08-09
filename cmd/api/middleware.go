package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pascaldekloe/jwt"
)

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // allow all requests
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		next.ServeHTTP(w, r)
	})
}

func (app *application) checkToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		// get token; authorization value from header
		authHeader := r.Header.Get("Authorization")

		// if authHeader == "" {
		// 	// could set an anonymous user
		// }

		// split the header by spaces
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			app.errorJSON(w, http.StatusBadRequest, errors.New("invalid auth header"))
			return
		}

		if headerParts[0] != "Bearer" {
			app.errorJSON(w, http.StatusUnauthorized, errors.New("unauthorized - no bearer"))
			return
		}

		// get token
		token := headerParts[1]

		// do hmac check
		claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secret))
		if err != nil {
			app.errorJSON(w, http.StatusForbidden, errors.New("unauthorized - failed hmac check"))
			return
		}

		// is token still valid at this time
		if !claims.Valid(time.Now()) {
			app.errorJSON(w, http.StatusForbidden, errors.New("unauthorized - token expired"))
			return
		}

		// check if audience is acceptable
		if !claims.AcceptAudience("mydomain.com") {
			app.errorJSON(w, http.StatusForbidden, errors.New("unauthorized - invalid audience"))
			return
		}

		// check issuer is your domain
		if claims.Issuer != "mydomain.com" {
			app.errorJSON(w, http.StatusForbidden, errors.New("unauthorized - invalid issuer"))
			return
		}

		// get user id from token
		userID, err := strconv.ParseInt(claims.Subject, 10, 64) // 64 bit
		if err != nil {
			app.errorJSON(w, http.StatusForbidden, errors.New("unauthorized"))
			return
		}

		app.logger.Println("Valid user:", userID)

		next.ServeHTTP(w, r)
	})
}
