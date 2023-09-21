package handler

import (
	"encoding/json"
	"net/http"
	"time"
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

func setJWTToken(w http.ResponseWriter, accessToken string, ttl time.Duration) {
	cookie := http.Cookie{
		Name:    "jwt_token",
		Value:   accessToken,
		Expires: time.Now().Add(ttl),
		Path:    "/api/auth",
	}

	http.SetCookie(w, &cookie)
}
