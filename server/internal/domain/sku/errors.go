package sku

import "errors"

var (
	ErrSKUNotFound      = errors.New("SKU not found")
	ErrInvalidSKUCode   = errors.New("SKU code cannot be empty")
	ErrInvalidSKUName   = errors.New("SKU name cannot be empty")
	ErrInvalidSKUPrice  = errors.New("SKU price must be positive")
	ErrInvalidSKUWeight = errors.New("SKU weight must be positive")
	ErrDuplicateSKUCode = errors.New("SKU code already exists")
)
