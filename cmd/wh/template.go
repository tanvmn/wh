package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
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
	PutawayBins      []data.PutAwayBin
	Page             any
	data.Item
	data.Purchase
	data.Account
	data.Receive
}

type PutawayPage struct {
	PutawayBins []data.PutAwayBin
	*data.Receive
}

type DifferenceActivitiesPage struct {
	DifferenceActivities []data.DifferenceActivity
}

type ReceiveProcessResultPage struct {
	*data.Receive
}

type PutawayResultPageTR struct {
	Quantity int
	Note     string
	*data.Bin
	*data.Item
}
type PutawayResultPage struct {
	*data.Receive
	TRs []*PutawayResultPageTR
}

func (ap *application) newPutawayResultPageByReceive(rc *data.Receive) (*PutawayResultPage, error) {
	if rc == nil {
		return nil, fmt.Errorf("parameter *Receive cannot be nil")
	}

	p := &PutawayResultPage{
		Receive: rc,
	}
	for _, iq := range p.Receive.Items {
		// add difference serials of the putaway by gtin
		ss, err := ap.data.DifferenceSerialsByGTINOfPutawayReceive(p.Receive.Purchase.Warehouse.ID, p.Receive.ID, iq.Item.GTIN)
		if err != nil {
			return nil, err
		}
		iq.Serials = append(iq.Serials, ss...)

		// accumilate the bin infos from all putaway serials (including the unsuccessful putaway ones)
		temp := make([]string, 0)
		for _, s := range iq.Serials {
			temp = append(temp, fmt.Sprintf("%v;%v;%v;%v", s.Bin.ID, s.Bin.Shelf, s.Bin.Row, s.Bin.Col))
		}
		bins := util.Set(temp...)

		// add a tr for each distinct bin (including an empty bin)
		for _, b := range bins {
			tr := &PutawayResultPageTR{
				Bin:  new(data.Bin),
				Item: &iq.Item,
				Note: iq.PutawayNote,
			}

			for _, s := range iq.Serials {
				if fmt.Sprintf("%v;%v;%v;%v", s.Bin.ID, s.Bin.Shelf, s.Bin.Row, s.Bin.Col) == b {
					tr.Quantity++
				}
			}

			if len(b) >= len("BIN-1;1;1;1") {
				split := strings.Split(b, ";")

				// add bin id
				tr.Bin.ID = split[0]

				// add bin shelf
				sh, err := strconv.ParseInt(split[1], 10, 64)
				if err != nil {
					return nil, err
				}
				tr.Bin.Shelf = sh

				// add bin row
				r, err := strconv.ParseInt(split[2], 10, 64)
				if err != nil {
					return nil, err
				}
				tr.Bin.Row = r

				// add bin col
				c, err := strconv.ParseInt(split[3], 10, 64)
				if err != nil {
					return nil, err
				}
				tr.Bin.Col = c
			}

			p.TRs = append(p.TRs, tr)
		}
	}

	return p, nil
}

func differenceActivityURL(id string) string {
	code := id[:4]
	switch code {
	case data.ReceiveIDCode:
		return fmt.Sprintf("%v/receive/%v/process/result", domain, id)
	case data.PutawayIDCode:
		return fmt.Sprintf("%v/putaway/%v/result", domain, data.ReceiveIDCode+id[4:])
	default:
		return domain + "/health"
	}
}

func actualPutawayQuantity(rc *data.Receive, gtin string) int {
	n := 0
	for _, iq := range rc.Items {
		if iq.Item.GTIN == gtin {
			for _, s := range iq.Serials {
				if len(s.Bin.ID) != 0 {
					n++
				}
			}
			break
		}
	}
	return n
}

func differenceActivityBadgeBg(idCode string) string {
	switch idCode[:4] {
	case data.ReceiveIDCode:
		return "warning"
	case data.PutawayIDCode:
		return "secondary"
	case data.PickIDCode:
		return "dark"
	case data.PackIDCode:
		return "danger"
	default:
		return "primary"
	}
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
	return strings.Contains(actualAt, "1000-01-01") || strings.Contains(actualAt, "01-01-1000")
}

func not01011000(time string) bool {
	return !notProcessed(time)
}

var fns = template.FuncMap{
	"badgeBg":                   badgeBg,
	"notProcessed":              notProcessed,
	"differenceActivityURL":     differenceActivityURL,
	"not01011000":               not01011000,
	"actualPutawayQuantity":     actualPutawayQuantity,
	"differenceActivityBadgeBg": differenceActivityBadgeBg,
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
