package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func renderJSON(w http.ResponseWriter, v interface{}) error {
	const op = "http-server.handler.renderJSON"

	js, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(js); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func getHeader(r *http.Request, header string) (string, error) {
	h := r.Header.Get(header)
	if h == "" {
		return "", fmt.Errorf(fmt.Sprintf("Header '%v' is missing", header))
	}

	return h, nil
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
