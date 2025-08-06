package main

import (
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

		i, err := ap.data.Item("8888021200126")
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, i, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	authenticate := middlewares{ap.sessionsManager.LoadAndSave, ap.authenticate}

	// File server
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", authenticate.then(http.StripPrefix("/rec", http.FileServerFS(rec.Files))))

	// Account
	mux.Handle("GET /account", authenticate.then(ap.account()))

	// Item
	mux.Handle("GET /items", authenticate.then(ap.items()))

	// Health
	mux.HandleFunc("GET /health", ap.health)

	// Home
	mux.Handle("GET /{$}", authenticate.then(ap.homePage()))

	// mux.Handle("GET /az", append(authenticate, ap.authorize("Kế toán trưởng")).then(ap.homePage()))

	// Login, logout
	mux.Handle("GET /login", authenticate.then(ap.loginPage()))
	mux.Handle("POST /login", authenticate.then(ap.login()))

	pre := middlewares{ap.recoverPanic, ap.logRequest, ap.addCommonHeaders}

	return pre.then(mux)
}
