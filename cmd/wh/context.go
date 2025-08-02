package main

type contextKey string

const (
	authenticatedID   = contextKey("authenticatedID")
	authenticatedRole = contextKey("authenticatedRole")
	authenticatedWarehouseID = contextKey("authenticatedRole")
	authenticatedStoreID = contextKey("authenticatedRole")
)
