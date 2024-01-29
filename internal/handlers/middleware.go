package handlers

import (
	"errors"
	"fmt"
	"imageAploaderS3/internal/helpers"
	"net/http"
)

func (m *Repository) AuthSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("AccessToken")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Redirect(w, r, "/signup", http.StatusSeeOther)
				return
			}
			http.Error(w, fmt.Sprintf("Error getting cookie: %v", err), http.StatusInternalServerError)
			return
		}
		if cookie.Value == "" {
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
		r.Header.Set("Authorization", cookie.Value)
		next.ServeHTTP(w, r)
	})
}

func (m *Repository) UserDataSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("TokenId")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Redirect(w, r, "/signup", http.StatusSeeOther)
				return
			}
			http.Error(w, fmt.Sprintf("Error getting cookie: %v", err), http.StatusInternalServerError)
			return
		}
		if cookie.Value == "" {
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
		res := helpers.GetJWTPayloadData(cookie.Value)
		m.App.Name = res.Name
		next.ServeHTTP(w, r)
	})
}
