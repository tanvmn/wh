package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"

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
		// 		// fmt.Println(pErr.SQLState())
		// 		// fmt.Println(pErr.Code.Class())
		// 		fmt.Println(pErr.Code.Name())
		// 	}
		// }

		// stmt := `select enum_range(null::type)`

		// var types []string
		// err := ap.data.DB.QueryRow(stmt).Scan(pq.Array(&types))

		f := struct {
			Bin []byte `json:"bin"`
		}{}

		bytes, err := fs.ReadFile(rec.Files, itemImgPathFS+"4983435734909.jpeg")
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		f.Bin = bytes

		err = ap.writeJSON(w, http.StatusOK, f, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/f", func(w http.ResponseWriter, r *http.Request) {
		o := struct {
			Bytes []byte `json:"bytes"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f, err := os.Create("./img.png")
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n, err := f.Write(o.Bytes)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%v bytes written\n", n)
	})

	authenticate := middlewares{ap.sessionsManager.LoadAndSave, ap.authenticate}

	// File server
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", authenticate.then(http.StripPrefix("/rec", http.FileServerFS(rec.Files))))

	// Account
	mux.Handle("GET /account", authenticate.then(ap.account()))

	// Item
	mux.Handle("GET /items", authenticate.then(ap.itemsPage()))
	mux.Handle("GET /items/json", authenticate.then(ap.items()))

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
