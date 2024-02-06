package handlers

import (
	"errors"
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
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
		m.App.Email = res.Email
		m.App.UserId = res.Sub
		next.ServeHTTP(w, r)
	})
}

func (m *Repository) XRayMiddleware(appName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			segName := appName + " - " + r.Method + " " + r.URL.Path
			ctx, seg := xray.BeginSegment(r.Context(), segName)

			defer func() {
				if seg != nil {
					seg.Close(nil)
				}
			}()

			if seg != nil {
				err := seg.AddAnnotation("UserName", m.App.Name)
				if err != nil {
					fmt.Println("adding annotation error: ", err)
				}
				err = seg.AddMetadata("UserInfo", map[string]string{
					"UserName": m.App.Name,
					"Email":    m.App.Email,
					"UserId":   m.App.UserId,
				})
				if err != nil {
					fmt.Println("adding metadata error: ", err)
				}
			}

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
