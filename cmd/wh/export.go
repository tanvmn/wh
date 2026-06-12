package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/util"
)

func (a *app) addExport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resupplyID := r.URL.Query().Get("resupply")

		exportID, err := a.data.AddExport(resupplyID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = a.data.SetResupplyStatus(resupplyID, data.AwaitingExport)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/export/%v", exportID), http.StatusSeeOther)
	})
}

func (a *app) exportPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		p := new(ExportPage)
		e, err := a.data.Export(id)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		p.Export = e

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "export", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) exportsByWarehousePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v; authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		es, err := a.data.ExportsByWarehouse(wID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// format the times to dd/mm/yyyy hh24:mi
		for i := range es {
			t, err := util.FormatRFC3339(es[i].ExpectedAt, util.DDMMYYYY24HMI)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].ExpectedAt = t

			t, err = util.FormatRFC3339(es[i].PickedAt, util.DDMMYYYY24HMI)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].PickedAt = t

			t, err = util.FormatRFC3339(es[i].PackedAt, util.DDMMYYYY24HMI)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].PackedAt = t

			ps, err := a.data.Packages(es[i].ID)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].Packages = ps
		}

		p := new(ExportsPage)
		p.Exports = es

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "exports", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) exportPickPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exportID := r.PathValue("id")

		picks, err := a.data.CalculatedPicks(exportID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		e, err := a.data.Export(exportID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ss, err := a.data.SerialsByWarehouse(e.Resupply.Account.Store.Warehouse.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ts, err := a.data.UnusedTotes(e.Resupply.Account.Store.Warehouse.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(ExportPickPage)
		p.Export = e
		p.Picks = picks
		p.Serials = ss
		p.UnusedTotes = ts

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "export_pick", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) pickExport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pickResult data.Export

		err := a.decodeJSON(w, r, &pickResult)
		if err != nil {
			a.log.Error(err.Error())
			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		ac, err := a.authenticatedAccount(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pickResult.PickedBy = *ac

		ex, err := a.data.Export(pickResult.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		rs, err := a.data.Resupply(ex.Resupply.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pickResult.Resupply = *rs

		err = a.data.PickExport(&pickResult)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/export/%v/pick/result", pickResult.ID), http.StatusSeeOther)
	})
}

func (a *app) exportPickResultPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		e, err := a.data.Export(id)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu xuất %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		p := new(ExportPickResultPage)
		p.Export = e

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "export_pick_result", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) exportPackPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		p := new(ExportPackPage)

		e, err := a.data.Export(id)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		// If the export is already packed
		if !util.Is01011000(e.PackedAt) {
			a.log.Error(fmt.Sprintf("Export %v is already packed at %v", e.ID, e.PackedAt))
			http.Error(w, fmt.Sprintf("Phiếu nhập %v đã đóng gói lúc %v", e.ID, e.PackedAt), http.StatusBadRequest)
			return
		}
		p.Export = e

		// Get the calcualted packages
		ps, err := a.data.CalculatedPackages(id)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.Packages = ps

		eI, err := data.ID64(id, data.ExportIDCode)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		ss, err := a.data.SerialsByExport(eI)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.Serials = ss

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "export_pack", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *app) packExport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var packResult data.Export

		err := ap.decodeJSON(w, r, &packResult)
		if err != nil {
			ap.log.Error(err.Error())

			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Add the account that packed the export
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.log.Error(fmt.Sprintf("%v; authenticatedCtxID: %v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		a, err := ap.data.Account(aID)
		if err != nil {
			ap.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		packResult.PackedBy = *a

		err = ap.data.PackExport(&packResult)
		if err != nil {
			ap.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// // Add the picked but unpacked serials to difference_serial table
		// err = ap.data.AddPackDifferenceSerials(packResult.ID)
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 	return
		// }

		// // Del the picked but unpacked serials from serial table
		// err = ap.data.DelUnpackedSerials(packResult.ID)
		// if err != nil {
		// 	ap.logger.Error(err.Error())
		// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		// 	return
		// }

		http.Redirect(w, r, fmt.Sprintf("/export/%v/pack/result", packResult.ID), http.StatusSeeOther)
	})
}

func (a *app) exportPackResultPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// id of the export
		id := r.PathValue("id")

		p := new(ExportPackResultPage)

		e, err := a.data.Export(id)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		p.Export = e

		ps, err := a.data.Packages(id)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.Packages = ps

		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = a.render(w, http.StatusOK, "export_pack_result", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) exportPackPromptPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = a.render(w, http.StatusOK, "export_pack_prompt", t)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) exportPackPageByPrompt() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nanoID := r.URL.Query().Get("serial")

		e, err := a.data.ExportByPickedSerial(nanoID)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập có Serial %v", nanoID), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/export/%v/pack", e.ID), http.StatusSeeOther)
	})
}
