package main

import (
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/rec"
	"github.com/tanNguyen2220022/wh/ui"
)

func (ap *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/t", func(w http.ResponseWriter, r *http.Request) {
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

		id, err := ap.data.Authenticate("0000000001", "pa55word")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(id)
	})

	protect := middlewares{ap.sessionsManager.LoadAndSave}

	// File server
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", protect.then(http.StripPrefix("/rec", http.FileServerFS(rec.Files))))

	// Public
	mux.HandleFunc("GET /health", ap.health)
	mux.Handle("GET /{$}", protect.then(ap.homePage()))

	// Login, logout
	mux.Handle("POST /login", protect.then(ap.login()))

	// Account
	mux.Handle("GET /account", protect.then(ap.account()))

	pre := middlewares{ap.recoverPanic, ap.logRequest, ap.addCommonHeaders}

	return pre.then(mux)
}
