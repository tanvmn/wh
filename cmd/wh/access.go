package main

import (
	"errors"
	"net/http"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/validator"
)

func (ap *application) loginPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := ap.render(w, http.StatusOK, "login", templData{}); err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

// login handles login form
func (ap *application) login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, "Không thể xử lý form đăng nhập", http.StatusBadRequest)
			return
		}

		phone := r.PostForm.Get("phone")
		password := r.PostForm.Get("password")

		va := new(validator.Validator)
		va.Check(validator.MinChars(phone, 10), "Số điện thoại tối thiểu 10 ký tự")
		va.Check(validator.MaxChars(phone, 12), "Số điện thoại tối đa 12 ký tự")
		va.Check(validator.NotBlank(password), "Mật khẩu không thể rỗng")

		if !va.Valid() {
			ap.logger.Error(va.Message())
			http.Error(w, va.Message(), http.StatusUnprocessableEntity)
			return
		}

		// i here is just the int64 part of account id
		id, err := ap.data.Authenticate(phone, password)
		if errors.Is(err, data.ErrInvalidCredentials) {
			ap.logger.Error(data.ErrInvalidCredentials.Error() + ", " + phone + ":" + password)
			http.Error(w, "Thông tin đăng nhập không chính xác", http.StatusUnprocessableEntity)
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

func (ap *application) logout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Renew the current session to change the session ID again
		err := ap.sessionsManager.RenewToken(r.Context())
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Remove the authenticatedID from the session data so the user is 'logged out'
		ap.sessionsManager.Remove(r.Context(), "authenticatedID")

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}
