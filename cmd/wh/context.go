package main

type contextKey string

const (
	authenticatedCtxID   = contextKey("authenticatedID")
	authenticatedCtxRole = contextKey("authenticatedRole")
	authenticatedCtxWarehouseID = contextKey("authenticatedRole")
	authenticatedCtxStoreID = contextKey("authenticatedRole")
)
