package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(js); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setCookies(w http.ResponseWriter, refreshToken string, accessToken string, refreshTokenTTL time.Duration, accessTokenTTL time.Duration) {
	httpOnlyCookie := http.Cookie{
		Name:     "httpOnly_cookie",
		Value:    refreshToken,
		Expires:  time.Now().Add(refreshTokenTTL),
		Path:     "/api/auth",
		HttpOnly: true,
	}
	http.SetCookie(w, &httpOnlyCookie)

	regularCookie := http.Cookie{
		Name:    "regular_cookie",
		Value:   accessToken,
		Expires: time.Now().Add(accessTokenTTL),
		Path:    "/api/auth",
	}
	http.SetCookie(w, &regularCookie)
}

func hashToken(token string) ([]byte, error) {
	const op = "hanlder.common.hashToken"

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", op, err)
	}

	return hashedToken, nil
}

func compareToken(providedToken string, hashedToken []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedToken, []byte(providedToken))

	return err == nil
}
