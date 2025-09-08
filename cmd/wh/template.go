package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/ui"
)

type templData struct {
	Domain           string
	CompanyName      string
	MinTimestamp     string
	Admin            string
	Accountant       string
	HeadAccountant   string
	Manager          string
	Employee         string
	AccountIDCode    string
	GTINIDCode       string
	SerialIDCode     string
	BinIDCode        string
	ToteIDCode       string
	BoxIDCode        string
	StaffIDCode      string
	WarehouseIDCode  string
	StoreIDCode      string
	SupplierIDCode   string
	PurchaseIDCode   string
	ReceiveIDCode    string
	ResupplyIDCode   string
	ExportIDCode     string
	AwaitingResponse string
	AwaitingReceive  string
	Receiving        string
	Ended            string
	Declined         string
	Items            []data.Item
	ItemQuantitys    []data.ItemQuantity
	Serials          []data.Serial
	Warehouses       []data.Warehouse
	Suppliers        []data.Supplier
	Purchases        []data.Purchase
	Receives         []data.Receive
	data.Item
	data.Purchase
	data.Account
	data.Receive
}

func badgeBg(status string) string {
	switch status {
	case data.AwaitingResponse:
		return "secondary"
	case data.AwaitingReceive:
		return "primary"
	case data.Receiving:
		return "warning"
	case data.Ended:
		return "success"
	case data.Declined:
		return "danger"
	default:
		return "dark"
	}
}

func notProcessed(actualAt string) bool {
	// return actualAt == "1000-01-01 00:00:00" || actualAt == "01-01-1000 00:00"
	return strings.Contains("1000-01-01 00:00:00", actualAt) || strings.Contains("01-01-1000 00:00", actualAt)
}

var fns = template.FuncMap{
	"badgeBg":   badgeBg,
	"notProcessed": notProcessed,
}

func (ap *application) newTemplData(r *http.Request) (templData, error) {
	aID, ok := r.Context().Value(authenticatedCtxID).(string)
	if !ok {
		return templData{}, fmt.Errorf("%w: account ID %v", ErrConvertCtxVal, aID)
	}
	wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
	if !ok {
		return templData{}, fmt.Errorf("%w: account's warehouse ID %v", ErrConvertCtxVal, wID)
	}

	ac, err := ap.data.Account(aID)
	if err != nil {
		return templData{}, err
	}

	wh, err := ap.data.Warehouse(wID)
	if !errors.Is(err, data.ErrNoWarehouses) && err != nil {
		return templData{}, err
	}
	if wh != nil {
		ac.Warehouse = *wh
	}

	ws, err := ap.data.Warehouses()
	if err != nil {
		return templData{}, err
	}

	ss, err := ap.data.Suppliers()
	if err != nil {
		return templData{}, err
	}

	return templData{
		Domain:           domain,
		CompanyName:      companyName,
		Admin:            data.Admin,
		Accountant:       data.Accountant,
		HeadAccountant:   data.HeadAccountant,
		Manager:          data.Manager,
		Employee:         data.Employee,
		AccountIDCode:    data.AccountIDCode,
		GTINIDCode:       data.GTINIDCode,
		SerialIDCode:     data.SerialIDCode,
		BinIDCode:        data.BinIDCode,
		ToteIDCode:       data.ToteIDCode,
		BoxIDCode:        data.BoxIDCode,
		StaffIDCode:      data.StaffIDCode,
		WarehouseIDCode:  data.WarehouseIDCode,
		StoreIDCode:      data.StoreIDCode,
		SupplierIDCode:   data.SupplierIDCode,
		PurchaseIDCode:   data.PurchaseIDCode,
		ReceiveIDCode:    data.ReceiveIDCode,
		ResupplyIDCode:   data.ResupplyIDCode,
		ExportIDCode:     data.ExportIDCode,
		AwaitingResponse: data.AwaitingResponse,
		AwaitingReceive:  data.AwaitingReceive,
		Receiving:        data.Receiving,
		Ended:            data.Ended,
		Declined:         data.Declined,
		Account:          *ac,
		Warehouses:       ws,
		Suppliers:        ss,
		MinTimestamp:     time.Now().Format(time.RFC3339)[:16],
	}, nil
}

func newTemplCache(lg *slog.Logger) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	// Get all the paths of the tmpl pages
	paths, err := fs.Glob(ui.Files, "html/pages/*.tmpl.html")
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}

	for _, path := range paths {
		// Get the *.tmpl.html part of the path, then the * part
		name := filepath.Base(path)
		name = name[:strings.Index(name, ".tmpl")]

		// Get the path's patterns of all tmpls needed for a page,
		// note that 'base' tmpl has to be the first element
		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.tmpl.html",
			path,
		}

		tmpl, err := template.New(name).Funcs(fns).ParseFS(ui.Files, patterns...)
		if err != nil {
			lg.Error(err.Error())
			return nil, err
		}

		cache[name] = tmpl
	}

	// Parse the login page
	cache["login"], err = template.ParseFS(ui.Files, "html/*.tmpl.html")
	if err != nil {
		lg.Error(err.Error())
		return nil, err
	}

	return cache, nil
}
