package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
)

func (ap *application) addExport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resupplyID := r.URL.Query().Get("resupply")

		exportID, err := ap.data.AddExport(resupplyID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.data.SetResupplyStatus(resupplyID, data.AwaitingExport)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Error(w, exportID, http.StatusBadRequest)
	})
}

func (ap *application) exportPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		p := new(ExportPage)
		e, err := ap.data.Export(id)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		p.Export = e

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "export", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) exportsByWarehousePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v; authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		es, err := ap.data.ExportsByWarehouse(wID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// format the times to dd/mm/yyyy hh24:mi
		for i := range es {
			t, err := util.FormatRFC3339(es[i].ExpectedAt, util.DDMMYYYY24HMI)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].ExpectedAt = t

			t, err = util.FormatRFC3339(es[i].PickedAt, util.DDMMYYYY24HMI)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].PickedAt = t

			t, err = util.FormatRFC3339(es[i].PackedAt, util.DDMMYYYY24HMI)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].PackedAt = t

			ps, err := ap.data.Packages(es[i].ID)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			es[i].Packages = ps
		}

		p := new(ExportsPage)
		p.Exports = es

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "exports", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) exportPickPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exportID := r.PathValue("id")

		picks, err := ap.data.CalculatedPicks(exportID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		e, err := ap.data.Export(exportID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ss, err := ap.data.SerialsByWarehouse(e.Resupply.Account.Store.Warehouse.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ts, err := ap.data.UnusedTotes(e.Resupply.Account.Store.Warehouse.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(ExportPickPage)
		p.Export = e
		p.Picks = picks
		p.Serials = ss
		p.UnusedTotes = ts

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "export_pick", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) pickExport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pickResult data.Export

		err := ap.decodeJSON(w, r, &pickResult)
		if err != nil {
			ap.logger.Error(err.Error())
			var mr *util.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		ac, err := ap.authenticatedAccount(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pickResult.PickedBy = *ac

		ex, err := ap.data.Export(pickResult.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		rs, err := ap.data.Resupply(ex.Resupply.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		pickResult.Resupply = *rs

		err = ap.data.PickExport(&pickResult)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/export/%v/pick/result", pickResult.ID), http.StatusSeeOther)
	})
}

func (ap *application) exportPickResultPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		e, err := ap.data.Export(id)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu xuất %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		p := new(ExportPickResultPage)
		p.Export = e

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "export_pick_result", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) exportPackPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		p := new(ExportPackPage)

		e, err := ap.data.Export(id)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		// If the export is already packed
		if !util.Is01011000(e.PackedAt) {
			ap.logger.Error(fmt.Sprintf("Export %v is already packed at %v", e.ID, e.PackedAt))
			http.Error(w, fmt.Sprintf("Phiếu nhập %v đã đóng gói lúc %v", e.ID, e.PackedAt), http.StatusBadRequest)
			return
		}
		p.Export = e

		// Get the calcualted packages
		ps, err := ap.data.CalculatedPackages(id)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.Packages = ps

		eI, err := data.ID64(id, data.ExportIDCode)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		ss, err := ap.data.SerialsByExport(eI)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.Serials = ss

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "export_pack", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) packExport() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var packResult data.Export

		err := ap.decodeJSON(w, r, &packResult)
		if err != nil {
			ap.logger.Error(err.Error())

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
			ap.logger.Error(fmt.Sprintf("%v; authenticatedCtxID: %v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		a, err := ap.data.Account(aID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		packResult.PackedBy = *a

		err = ap.data.PackExport(&packResult)
		if err != nil {
			ap.logger.Error(err.Error())
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

func (ap *application) exportPackResultPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// id of the export
		id := r.PathValue("id")

		p := new(ExportPackResultPage)

		e, err := ap.data.Export(id)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoExports) {
				http.Error(w, fmt.Sprintf("Không tìm thấy phiếu nhập %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		p.Export = e

		ps, err := ap.data.Packages(id)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		p.Packages = ps

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "export_pack_result", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) exportPackPromptPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = ap.render(w, http.StatusOK, "export_pack_prompt", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) exportPackPageByPrompt() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nanoID := r.URL.Query().Get("serial")

		e, err := ap.data.ExportByPickedSerial(nanoID)
		if err != nil {
			ap.logger.Error(err.Error())
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
