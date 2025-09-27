package main

import (
	"fmt"
	"net/http"

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
		// 		fmt.Println(pErr.Code.Name())
		// 		fmt.Println(pErr.Message)
		// 		fmt.Println(pErr.SQLState())
		// 		fmt.Println(pErr.Code.Class())
		// 	}
		// }

		// o := struct {
		// 	Bytes []byte `json:"bytes"`
		// }{}

		// err := json.NewDecoder(r.Body).Decode(&o)
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// f, err := os.Create("./img.png")
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// n, err := f.Write(o.Bytes)
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// s, err := ap.data.UnsafeStocks("WAR-1")
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// println()
		// println("unsafe stock")
		// for _, iq := range s {
		// 		println(iq.Item.GTIN, "stock", iq.Stock, "restock", iq.Restock)
		// }

		ps, err := ap.data.Purchases("WAR-1")
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, p := range ps {
			println(p.ID, p.Status)
		}
	})

	identify := middlewares{ap.sessionsManager.LoadAndSave, ap.identify}

	// File server
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", identify.then(http.StripPrefix("/rec", http.FileServerFS(rec.Files))))

	// Health
	mux.HandleFunc("GET /health", ap.health)

	// Login, logout
	mux.Handle("GET /login", identify.then(ap.loginPage()))
	mux.Handle("GET /logout", identify.then(ap.logout()))
	mux.Handle("POST /login", identify.then(ap.login()))

	// Account
	mux.Handle("GET /account/{id}", identify.then(ap.account()))

	// Item
	mux.Handle("GET /items", identify.then(ap.itemsPage()))
	mux.Handle("GET /items/json", identify.then(ap.items()))
	mux.Handle("GET /items-by-supplier/json", identify.then(ap.itemsBySupplier()))
	mux.Handle("GET /unsafe-stocks", identify.then(ap.unsafeStocksPage()))

	// Serial
	mux.Handle("GET /serials", identify.then(ap.serialsPage()))

	// Warehouse
	mux.Handle("GET /totes/{warehouse}/unused/json", identify.then(ap.unusedTotes()))

	// Supplier
	mux.Handle("GET /suppliers/json", identify.then(ap.suppliers()))

	// Home
	mux.Handle("GET /{$}", identify.then(ap.homePage()))

	// Purchase
	mux.Handle("GET /purchase/{id}", append(identify, ap.permit(data.Accountant)).then(ap.purchasePage()))
	mux.Handle("GET /purchase/{id}/json", append(identify, ap.permit(data.Accountant)).then(ap.purchase()))
	mux.Handle("GET /purchases", append(identify, ap.permit(data.Accountant)).then(ap.purchasesPage()))
	mux.Handle("GET /add-purchase", append(identify, ap.permit(data.Accountant)).then(ap.addPurchasePage()))
	mux.Handle("POST /purchase", append(identify, ap.permit(data.Accountant)).then(ap.addPurchase()))
	mux.Handle("PUT /purchase", append(identify, ap.permit(data.Accountant)).then(ap.setPurchase()))
	mux.Handle("DELETE /purchase/{id}", append(identify, ap.permit(data.Accountant)).then(ap.delPurchase()))

	// Receive
	mux.Handle("GET /add-receive", append(identify, ap.permit(data.Accountant)).then(ap.addReceivePage()))
	mux.Handle("GET /receive/{id}", append(identify, ap.permit(data.Accountant, data.Manager, data.Employee)).then(ap.receivePage()))
	mux.Handle("GET /receive/{id}/json", append(identify, ap.permit(data.Accountant, data.Manager, data.Employee)).then(ap.receive()))
	mux.Handle("GET /receive/{id}/process", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.receiveProcessPage()))
	mux.Handle("GET /receive/{id}/process/result", append(identify, ap.permit(data.Manager, data.Employee, data.Accountant)).then(ap.receiveProcessResultPage()))
	mux.Handle("GET /receives", append(identify, ap.permit(data.Accountant, data.Manager, data.Employee)).then(ap.receivesPage()))
	mux.Handle("GET /receives-by-purchase/{purchase}", append(identify, ap.permit(data.Accountant, data.Manager, data.Employee)).then(ap.receivesByPurchasePage()))
	mux.Handle("POST /receive", append(identify, ap.permit(data.Accountant)).then(ap.addReceive()))
	mux.Handle("POST /receive/process", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.processReceive()))
	mux.Handle("PUT /receive", append(identify, ap.permit(data.Accountant)).then(ap.setReceive()))
	mux.Handle("DELETE /receive/{id}", append(identify, ap.permit(data.Accountant)).then(ap.delReceive()))

	// Putaway
	mux.Handle("GET /putaway-prompt", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.putawayPromptPage()))
	mux.Handle("GET /putaway", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.putawayPageBySerial()))
	mux.Handle("GET /putaway/{receive}", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.putawayPage()))
	mux.Handle("GET /putaway/{receive}/result", append(identify, ap.permit(data.Manager, data.Employee, data.Accountant)).then(ap.putawayResultPage()))
	mux.Handle("POST /putaway", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.putaway()))

	// Resuppy
	mux.Handle("GET /add-resupply", append(identify, ap.permit(data.Manager, data.Employee), ap.permitStoreEmployee).then(ap.resupplyAddPage()))
	mux.Handle("GET /resupply/{id}", append(identify, ap.permit(data.Manager, data.Employee), ap.permitStoreEmployee).then(ap.resupplyPage()))
	mux.Handle("GET /resupplies", append(identify, ap.permit(data.Manager, data.Employee), ap.permitStoreEmployee).then(ap.resuppliesPage()))
	mux.Handle("POST /resupply", append(identify, ap.permit(data.Manager, data.Employee), ap.permitStoreEmployee).then(ap.addResupply()))
	mux.Handle("PUT /resupply", append(identify, ap.permit(data.Manager, data.Employee), ap.permitStoreEmployee).then(ap.setResupply()))
	mux.Handle("PUT /resupply/decline", append(identify, ap.permit(data.Manager, data.Employee), ap.permitStoreEmployee).then(ap.declineResupply()))
	mux.Handle("DELETE /resupply/{id}", append(identify, ap.permit(data.Manager, data.Employee), ap.permitStoreEmployee).then(ap.delResupply()))

	// Export
	mux.Handle("GET /export/{id}", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportPage()))
	mux.Handle("GET /export/{id}/pick", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportPickPage()))
	mux.Handle("GET /export/{id}/pick/result", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportPickResultPage()))
	mux.Handle("GET /export/{id}/pack", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportPackPage()))
	mux.Handle("GET /export/{id}/pack/result", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportPackResultPage()))
	mux.Handle("GET /exports", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportsByWarehousePage()))
	mux.Handle("GET /pack-prompt", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportPackPromptPage()))
	mux.Handle("GET /pack-prompt/process", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.exportPackPageByPrompt()))
	mux.Handle("POST /export", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.addExport()))
	mux.Handle("POST /export/pick", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.pickExport()))
	mux.Handle("POST /export/pack", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.packExport()))

	// Difference Activities
	mux.Handle("GET /difference-activities", append(identify, ap.permit(data.Manager, data.Employee)).then(ap.differenceActivitiesPage()))

	pre := middlewares{ap.recoverPanic, ap.logRequest, ap.addHeaders}

	return pre.then(mux)
}
