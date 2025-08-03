package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/rec"
	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /t", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../fe/index.html")
	})
	mux.HandleFunc("POST /t", func(w http.ResponseWriter, r *http.Request) {
		// stmt := `insert into account (phone) values (0000000001)`
		// _, err := ap.data.DB.Exec(stmt)
		// if err != nil {
		// 	var pErr *pq.Error
		// 	if errors.As(err, &pErr) {
		// 		fmt.Printf("%+v\n", pErr)
		// 		fmt.Println(pErr.Code)
		// 		fmt.Println(pErr.Message)
		// 		fmt.Println(pErr.SQLState())
		// 		fmt.Println(pErr.Code.Class())
		// 		fmt.Println(pErr.Code.Name())
		// 	}
		// }
		payload := []struct {
			I int    `json:"i,omitzero,omitempty"`
			S string `json:"s,omitzero,omitempty"`
			A []int  `json:"a,omitzero,omitempty"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			ap.logger.Error(err.Error())
		}

		fmt.Printf("%+v\n", payload)

		err = ap.writeJSON(w, http.StatusOK, payload, nil)
		if err != nil {
			ap.logger.Error(err.Error())
		}
	})

	authenticate := middlewares{ap.sessionsManager.LoadAndSave, ap.authenticate}

	// File server
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", authenticate.then(http.StripPrefix("/rec", http.FileServerFS(rec.Files))))

	// Health
	mux.HandleFunc("GET /health", ap.health)

	// Home
	mux.Handle("GET /{$}", authenticate.then(ap.homePage()))

	// mux.Handle("GET /az", append(authenticate, ap.authorize("Kế toán trưởng")).then(ap.homePage()))

	// Login, logout
	mux.Handle("GET /login", authenticate.then(ap.loginPage()))
	mux.Handle("POST /login", authenticate.then(ap.login()))

	// Account
	mux.Handle("GET /account", authenticate.then(ap.account()))

	pre := middlewares{ap.recoverPanic, ap.logRequest, ap.addCommonHeaders}

	return pre.then(mux)
}
