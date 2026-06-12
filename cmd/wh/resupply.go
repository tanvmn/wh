package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/util"
)

func (a *app) resupplyAddPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sI, ok := r.Context().Value(authenticatedCtxStoreID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v; authenticatedCtxStoreID %v", ErrConvertCtxVal, sI))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s, err := a.data.Store(sI)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(ResupplyAddPage)
		stocks, err := a.data.StocksByWarehouse(s.Warehouse.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(stocks) == 0 {
			msg := fmt.Sprintf("Hiện kho %v/%v không còn hàng để cung cấp", s.Warehouse.ID, s.Warehouse.Name)
			a.log.Error(msg)
			http.Error(w, msg, http.StatusUnprocessableEntity)
			return
		}
		p.Stocks = stocks

		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = a.render(w, http.StatusOK, "resupply_add", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) addResupply() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rs data.Resupply

		err := a.decodeJSON(w, r, &rs)
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

		// Get and set the resupply's account, store data
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v; authenticatedCtxID %v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		ac, err := a.data.Account(aID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		st, err := a.data.Store(ac.Store.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		ac.Store = *st
		rs.Account = *ac

		// Add the resupply
		rID, err := a.data.AddResupply(&rs)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%v/resupply/%v", domain, rID), http.StatusSeeOther)
	})
}

func (a *app) resupplyPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		rs, err := a.data.Resupply(id)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoResupplies) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu xuất %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// Set the max quantity an item of resupply can be set to
		err = a.data.SetMaxResupplyItemQuantities(rs)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Start removing the resupply items from the warehouse remaining stocks for the datalist options
		stocks, err := a.data.StocksByWarehouse(rs.Account.Store.Warehouse.ID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(ResupplyPage)
		p.Resupply = rs

		for _, s := range stocks {
			contained := false
			for _, iq := range rs.Items {
				if s.Item.GTIN == iq.Item.GTIN {
					contained = true
					break
				}
			}

			if !contained {
				p.ItemOpts = append(p.ItemOpts, s)
			}
		}

		// Prepare the template data
		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = a.render(w, http.StatusOK, "resupply", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) setResupply() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rs data.Resupply

		err := a.decodeJSON(w, r, &rs)
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

		err = a.data.SetResupply(&rs)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%v/resupply/%v", domain, rs.ID), http.StatusSeeOther)
	})
}

func (a *app) delResupply() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		err := a.data.DelResupply(id)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%v/resupplies", domain), http.StatusSeeOther)
	})
}

func (a *app) resuppliesPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			a.log.Error(fmt.Sprintf("%v; authenticatedCtxID: %v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ac, err := a.data.Account(aID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var warehouseID string
		if len(ac.Warehouse.ID) > 4 {
			warehouseID = ac.Warehouse.ID
		} else {
			st, err := a.data.Store(ac.Store.ID)
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			warehouseID = st.Warehouse.ID
		}

		p := new(ResuppliesPage)

		rs, err := a.data.ResuppliesByWarehouse(warehouseID)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		for i := range rs {
			rs[i].ExpectedAt, err = util.FormatRFC3339(rs[i].ExpectedAt, "02-01-2006 15:04")
			if err != nil {
				a.log.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		p.Resupplies = rs

		td, err := a.newTemplData(r)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = a.render(w, http.StatusOK, "resupplies", td)
		if err != nil {
			a.log.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (a *app) declineResupply() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rs data.Resupply

		err := a.decodeJSON(w, r, &rs)
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

		err = a.data.DeclineResupply(&rs)
		if err != nil {
			a.log.Error(err.Error())
			if errors.Is(err, data.ErrNoResupplies) {
				http.Error(w, fmt.Sprintf("Không tìm thấy yêu cầu xuất %v", rs.ID), http.StatusBadRequest)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%v/resupply/%v", domain, rs.ID), http.StatusSeeOther)
	})
}
