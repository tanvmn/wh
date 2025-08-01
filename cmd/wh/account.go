package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
	"github.com/tanNguyen2220022/wh/internal/validator"
)

// account handles GET /account?id= and response a JSON of internal/data.Account
func (ap *application) account() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sID := r.URL.Query().Get("id")

		id, err := ap.parseID(sID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ac, err := ap.data.Account(id)
		if errors.Is(err, data.ErrNoAccounts) {
			s := fmt.Sprintf("không tìm thấy tài khoản, ACC-%v", id)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusNotFound)
			return
		} else if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ac.BDate, err = util.FormatRFC3339(ac.BDate, time.DateOnly)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, ac, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

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
