package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/validator"
	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) loginPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// path := filepath.Join()
		// http.ServeFile(w,r,)
		http.ServeFileFS(w, r, ui.Files, "html/login.html")
	})
}

// login handles login by form
func (ap *application) login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		phone := r.PostForm.Get("phone")
		password := r.PostForm.Get("password")

		va := new(validator.Validator)
		va.Check(validator.MinChars(phone, 10), "Số điện thoại tối thiểu 10 ký tự")
		va.Check(validator.NotBlank(password), "Mật khẩu không thể rỗng")

		if !va.Valid() {
			ap.logger.Error(va.Message())
			http.Error(w, va.Message(), http.StatusBadRequest)
			return
		}

		id, err := ap.data.Authenticate(phone, password)
		if errors.Is(err, data.ErrInvalidCredentials) {
			s := fmt.Sprintf("%v:%v\nThông tin đăng nhập không chính xác", phone, password)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusUnprocessableEntity)
			return
		} else if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Renew token everytime authentication status or permission changes is good practice
		err = ap.sessionsManager.RenewToken(r.Context())
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ap.sessionsManager.Put(r.Context(), "authenticatedID", id)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
