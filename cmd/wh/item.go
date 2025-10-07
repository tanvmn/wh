package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tanNguyen2220022/wh/internal/data"
	"github.com/tanNguyen2220022/wh/internal/util"
)

func (ap *application) name(i data.Item) string {
	return fmt.Sprintf("%v, %v, màu %v, %v, %v", i.Type, i.Brand, i.Color, i.Size, i.Characteristic)
}

func (ap *application) itemsPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		is, err := ap.data.Items()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		for i := range is {
			is[i].Stock, err = ap.data.Stock(is[i].GTIN, wID)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		if is != nil {
			td.Items = is
		}

		err = ap.render(w, http.StatusOK, "items", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) items() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		is, err := ap.data.Items()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if is == nil {
			ap.logger.Error(data.ErrNoItems.Error())
			http.Error(w, "Không tìm thấy mặt hàng nào", http.StatusNotFound)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, is, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) itemsBySupplier() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sID := r.URL.Query().Get("supplier")

		sp, err := ap.data.Supplier(sID)
		if err != nil {
			ap.logger.Error(err.Error())

			if errors.Is(err, data.ErrNoSuppliers) {
				http.Error(w, fmt.Sprintf("Không tìm thấy kho hoặc kho không tồn tại, ID: %v", sID), http.StatusBadRequest)
			} else {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		is, err := ap.data.ItemsBySupplier(sp.ID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if is == nil {
			s := fmt.Sprintf("Không tìm thấy hàng theo nhà cung cấp %v", sp.ID)
			ap.logger.Error(s)
			http.Error(w, s, http.StatusNotFound)
			return
		}

		err = ap.writeJSON(w, http.StatusOK, is, nil)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) unsafeStocksPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		s, err := ap.data.UnsafeStocks(wID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if len(s) == 0 {
			http.Error(w, fmt.Sprintf("Kho %v hiện không có hàng dưới lượng an toàn", wID), http.StatusNotFound)
			return
		}

		p := new(UnsafeStockPage)
		p.UnsafeStocks = s

		td, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		td.Page = p

		err = ap.render(w, http.StatusOK, "unsafe_stock", td)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) addUnsafePurchases() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ExpectedAt string              `json:"expectedAt,omitempty,omitzero"`
			Items      []data.ItemQuantity `json:"items,omitempty,omitzero"`
		}

		err := ap.decodeJSON(w, r, &req)
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

		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		aID, ok := r.Context().Value(authenticatedCtxID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v; authenticatedCtxID: %v", ErrConvertCtxVal, aID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Get the distinct supplier ids
		sups := []string{}
		for _, iq := range req.Items {
			sups = append(sups, iq.Supplier.ID)
		}
		sups = util.Set(sups...)

		// Prepare the *data.Purchase and add
		for _, s := range sups {
			p := new(data.Purchase)
			p.Warehouse.ID = wID
			p.Account.ID = aID
			p.Supplier.ID = s
			p.ExpectedAt = req.ExpectedAt

			for _, iq := range req.Items {
				if iq.Supplier.ID == s {
					iq.Quantity = iq.Restock
					p.Items = append(p.Items, iq)
				}
			}

			// Check against warehouse capacity
			enough, err := ap.data.CheckCapacity(p.Items, p.Warehouse.ID)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if !enough {
				ap.logger.Error("Not enough capacity")
				http.Error(w, fmt.Sprintf("Kho %v hiện không đủ sức chứa", p.Warehouse.ID), http.StatusUnprocessableEntity)
				return
			}

			_, _, err = ap.data.AddPurchase(p)
			if err != nil {
				ap.logger.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/purchases", http.StatusSeeOther)
	})
}

func (ap *application) outOfDateItems() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wID, ok := r.Context().Value(authenticatedCtxWarehouseID).(string)
		if !ok {
			ap.logger.Error(fmt.Sprintf("%v, authenticatedCtxWarehouseID: %v", ErrConvertCtxVal, wID))
		}

		is, err := ap.data.OutOfDateItems(wID)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(OutOfDateItemsPage)
		p.Items = is

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "outofdate_items", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) itemAddPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is, err := ap.data.AllItems()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ss, err := ap.data.Suppliers()
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		p := new(ItemAddPage)
		p.Items = is
		p.Suppliers = ss

		t, err := ap.newTemplData(r)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Page = p

		err = ap.render(w, http.StatusOK, "item_add", t)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})
}

func (ap *application) addItem() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, fheader, err := r.FormFile("img")
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() {
			if err2 := file.Close(); err2 != nil {
				panic(err2)
			}
		}()

		// read the bytes from form file to a buf and use that to detect the content
		buf := make([]byte, fheader.Size)
		_, err = file.Read(buf)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ftype := strings.Split(http.DetectContentType(buf), "/")[1]

		// create the image file
		path := filepath.Join(".", "rec", "item", "img", fmt.Sprintf("%v.%v", r.PostForm.Get("gtin"), ftype))
		fOut, err := os.Create(path)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err3 := fOut.Close(); err3 != nil {
				panic(err3)
			}
		}()

		_, err = fOut.Write(buf)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// initialize an *Item for adding to db
		i := new(data.Item)
		i.GTIN = r.PostForm.Get("gtin")
		i.Type = r.PostForm.Get("type")

		i.Volume, err = strconv.ParseFloat(r.PostForm.Get("volume"), 64)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f64, err := strconv.ParseFloat(r.PostForm.Get("weight"), 64)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		i.Weight = int64(f64)

		i.Brand = r.PostForm.Get("brand")
		i.Material = r.PostForm.Get("material")
		i.Color = r.PostForm.Get("color")
		i.Size = r.PostForm.Get("size")

		f64, err = strconv.ParseFloat(r.PostForm.Get("shelfLife"), 64)
		if err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		i.ShelfLife = int64(f64)

		i.Characteristic = r.PostForm.Get("characteristic")
		i.ImgFSPath = filepath.Join("item", "img", fmt.Sprintf("%v.%v", r.PostForm.Get("gtin"), ftype))
		i.Supplier.ID = r.PostForm.Get("supplier")

		// add the item
		if err = ap.data.AddItem(i); err != nil {
			ap.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		time.Sleep(1600 * time.Millisecond)
		http.Redirect(w, r, "/items", http.StatusSeeOther)
	})
}
