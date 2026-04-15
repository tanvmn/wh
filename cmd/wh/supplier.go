package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanvmn/wh/internal/data"
	"github.com/tanvmn/wh/internal/util"
)

func (ap *application) suppliers() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ss, err := ap.data.Suppliers()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if ss == nil {
			ap.logger.Error(data.ErrNoSuppliers.Error())
			http.Error(w, "Không tìm thấy nhà cung cấp nào", http.StatusNotFound)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, ss, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) addSupplierPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is, err := ap.data.AllItems()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(SupplierAddPage)
		p.Items = is

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "supplier_add", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) addSupplier() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input data.Supplier

		err := ap.decodeJSON(w, r, &input)
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

		id, err := ap.data.AddSupplier(&input)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/supplier/%v", id), http.StatusSeeOther)
	})
}

func (ap *application) supplierPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		s, err := ap.data.Supplier(id)
		if err != nil {
			ap.logger.Error(err.Error())
			if errors.Is(err, data.ErrNoSuppliers) {
				http.Error(w, fmt.Sprintf("Không tìm thấy nhà cung cấp %v", id), http.StatusNotFound)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		p := new(SupplierPage)
		p.Supplier = s

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "supplier", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) suppliersPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ss, err := ap.data.Suppliers()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(SuppliersPage)
		p.Suppliers = ss

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "suppliers", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}
