package main

import (
	"net/http"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/rec"
	"github.com/tanvmn/wh/ui"
)

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	identify := middlewares{a.sessionsManager.LoadAndSave, a.identify}

	// File server
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.Handle("GET /rec/", identify.then(http.StripPrefix("/rec", http.FileServerFS(rec.Files))))

	// Health
	mux.HandleFunc("GET /health", a.health)

	// Login, logout
	mux.Handle("GET /login", identify.then(a.loginPage()))
	mux.Handle("GET /logout", identify.then(a.logout()))
	mux.Handle("POST /login", identify.then(a.login()))

	// Account
	mux.Handle("GET /account/{id}", identify.then(a.account()))

	// Item
	mux.Handle("GET /items", identify.then(a.itemsPage()))
	mux.Handle("GET /items/json", identify.then(a.items()))
	mux.Handle("GET /items-by-supplier/json", identify.then(a.itemsBySupplier()))
	mux.Handle("GET /items/out-of-date", identify.then(a.outOfDateItems()))
	mux.Handle("GET /item/add", identify.then(a.itemAddPage()))
	mux.Handle("POST /item/add", identify.then(a.addItem()))

	// Supplier
	mux.Handle("GET /supplier/add", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.addSupplierPage()))
	mux.Handle("GET /supplier/{id}", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.supplierPage()))
	mux.Handle("GET /suppliers", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.suppliersPage()))
	mux.Handle("POST /supplier/add", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.addSupplier()))

	// Unsafe
	mux.Handle("GET /unsafe-stocks", identify.then(a.unsafeStocksPage()))
	mux.Handle("POST /unsafe/purchases", identify.then(a.addUnsafePurchases()))

	// Serial
	mux.Handle("GET /serials", identify.then(a.serialsPage()))
	mux.Handle("GET /serials/out-of-date", identify.then(a.outOfDateSerialsPage()))
	mux.Handle("GET /serials-by-bin", identify.then(a.serialsByBinPage()))

	// Warehouse
	mux.Handle("GET /totes/{warehouse}/unused/json", identify.then(a.unusedTotes()))
	mux.Handle("GET /bins", identify.then(a.binsPage()))

	// Supplier
	mux.Handle("GET /suppliers/json", identify.then(a.suppliers()))

	// Home
	mux.Handle("GET /{$}", identify.then(a.homePage()))

	// Purchase
	mux.Handle("GET /purchase/{id}", append(identify, a.permit(data.Accountant)).then(a.purchasePage()))
	mux.Handle("GET /purchase/{id}/json", append(identify, a.permit(data.Accountant)).then(a.purchase()))
	mux.Handle("GET /purchases", append(identify, a.permit(data.Accountant)).then(a.purchasesPage()))
	mux.Handle("GET /add-purchase", append(identify, a.permit(data.Accountant)).then(a.addPurchasePage()))
	mux.Handle("POST /purchase", append(identify, a.permit(data.Accountant)).then(a.addPurchase()))
	mux.Handle("PUT /purchase", append(identify, a.permit(data.Accountant)).then(a.setPurchase()))
	mux.Handle("DELETE /purchase/{id}", append(identify, a.permit(data.Accountant)).then(a.delPurchase()))

	// Receive
	mux.Handle("GET /add-receive", append(identify, a.permit(data.Accountant)).then(a.addReceivePage()))
	mux.Handle("GET /receive/{id}", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.receivePage()))
	mux.Handle("GET /receive/{id}/json", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.receive()))
	mux.Handle("GET /receive/{id}/process", append(identify, a.permit(data.Manager, data.Employee)).then(a.receiveProcessPage()))
	mux.Handle("GET /receive/{id}/process/result", append(identify, a.permit(data.Manager, data.Employee, data.Accountant)).then(a.receiveProcessResultPage()))
	mux.Handle("GET /receives", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.receivesPage()))
	mux.Handle("GET /receives-by-purchase/{purchase}", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.receivesByPurchasePage()))
	mux.Handle("POST /receive", append(identify, a.permit(data.Accountant)).then(a.addReceive()))
	mux.Handle("POST /receive/process", append(identify, a.permit(data.Manager, data.Employee)).then(a.processReceive()))
	mux.Handle("PUT /receive", append(identify, a.permit(data.Accountant)).then(a.setReceive()))
	mux.Handle("DELETE /receive/{id}", append(identify, a.permit(data.Accountant)).then(a.delReceive()))

	// Putaway
	mux.Handle("GET /putaway-prompt", append(identify, a.permit(data.Manager, data.Employee)).then(a.putawayPromptPage()))
	mux.Handle("GET /putaway", append(identify, a.permit(data.Manager, data.Employee)).then(a.putawayPageBySerial()))
	mux.Handle("GET /putaway/{receive}", append(identify, a.permit(data.Manager, data.Employee)).then(a.putawayPage()))
	mux.Handle("GET /putaway/{receive}/result", append(identify, a.permit(data.Manager, data.Employee, data.Accountant)).then(a.putawayResultPage()))
	mux.Handle("POST /putaway", append(identify, a.permit(data.Manager, data.Employee)).then(a.putaway()))

	// Resuppy
	mux.Handle("GET /add-resupply", append(identify, a.permit(data.Manager, data.Employee), a.permitStoreEmployee).then(a.resupplyAddPage()))
	mux.Handle("GET /resupply/{id}", append(identify, a.permit(data.Manager, data.Employee), a.permitStoreEmployee).then(a.resupplyPage()))
	mux.Handle("GET /resupplies", append(identify, a.permit(data.Manager, data.Employee), a.permitStoreEmployee).then(a.resuppliesPage()))
	mux.Handle("POST /resupply", append(identify, a.permit(data.Manager, data.Employee), a.permitStoreEmployee).then(a.addResupply()))
	mux.Handle("PUT /resupply", append(identify, a.permit(data.Manager, data.Employee), a.permitStoreEmployee).then(a.setResupply()))
	mux.Handle("PUT /resupply/decline", append(identify, a.permit(data.Manager, data.Employee), a.permitStoreEmployee).then(a.declineResupply()))
	mux.Handle("DELETE /resupply/{id}", append(identify, a.permit(data.Manager, data.Employee), a.permitStoreEmployee).then(a.delResupply()))

	// Export
	mux.Handle("GET /export/{id}", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportPage()))
	mux.Handle("GET /export/{id}/pick", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportPickPage()))
	mux.Handle("GET /export/{id}/pick/result", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportPickResultPage()))
	mux.Handle("GET /export/{id}/pack", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportPackPage()))
	mux.Handle("GET /export/{id}/pack/result", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportPackResultPage()))
	mux.Handle("GET /exports", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportsByWarehousePage()))
	mux.Handle("GET /pack-prompt", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportPackPromptPage()))
	mux.Handle("GET /pack-prompt/process", append(identify, a.permit(data.Manager, data.Employee)).then(a.exportPackPageByPrompt()))
	mux.Handle("POST /export", append(identify, a.permit(data.Manager, data.Employee)).then(a.addExport()))
	mux.Handle("POST /export/pick", append(identify, a.permit(data.Manager, data.Employee)).then(a.pickExport()))
	mux.Handle("POST /export/pack", append(identify, a.permit(data.Manager, data.Employee)).then(a.packExport()))

	// Inventory
	mux.Handle("GET /inventory-add", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.addInventoryPage()))
	mux.Handle("GET /inventories", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.inventoriesPage()))
	mux.Handle("GET /inventory/{id}", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.inventoryPage()))
	mux.Handle("GET /inventory/{id}/process", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.inventoryProcessPage()))
	mux.Handle("GET /inventory/{id}/process/result", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.inventoryProcessResultPage()))
	mux.Handle("POST /inventory", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.addInventory()))
	mux.Handle("POST /inventory/{id}/bin-result", append(identify, a.permit(data.Accountant, data.Manager, data.Employee)).then(a.processInventoryBinResult()))

	// Difference Activities
	mux.Handle("GET /difference-activities", append(identify, a.permit(data.Manager, data.Employee)).then(a.differenceActivitiesPage()))

	pre := middlewares{a.limitRate, a.recoverPanic, a.logRequest, a.addHeaders}

	return pre.then(mux)
}
