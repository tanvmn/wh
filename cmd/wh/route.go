package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/tanNguyen2220022/wh/internal/data"
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

		ss, err := ap.data.Serials("4983435734503")
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%+v", ss)
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

	identify := middlewares{ap.sessionsManager.LoadAndSave, ap.identify}

	// File server
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", identify.then(http.StripPrefix("/rec", http.FileServerFS(rec.Files))))

	// Health
	mux.HandleFunc("GET /health", ap.health)

	// Login, logout
	mux.Handle("GET /login", identify.then(ap.loginPage()))
	mux.Handle("POST /login", identify.then(ap.login()))

	// Account
	mux.Handle("GET /account", identify.then(ap.account()))

	// Item
	mux.Handle("GET /items", identify.then(ap.itemsPage()))
	mux.Handle("GET /items/json", identify.then(ap.items()))
	mux.Handle("GET /serials", identify.then(ap.serialsPage()))
	mux.Handle("GET /items-by-supplier", identify.then(ap.itemsBySupplier()))

	// Supplier
	mux.Handle("GET /suppliers/json", identify.then(ap.suppliers()))

	// Home
	mux.Handle("GET /{$}", identify.then(ap.homePage()))

	// Purchase
	mux.Handle("GET /purchase/{id}", append(identify, ap.permit(data.Accountant, data.HeadAccountant)).then(ap.purchasePage()))
	mux.Handle("GET /add-purchase", append(identify, ap.permit(data.Accountant, data.HeadAccountant)).then(ap.addPurchasePage()))
	mux.Handle("POST /purchase", append(identify, ap.permit(data.Accountant, data.HeadAccountant)).then(ap.addPurchase()))

	pre := middlewares{ap.recoverPanic, ap.logRequest, ap.addHeaders}

	return pre.then(mux)
}
