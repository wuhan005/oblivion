// Copyright 2021 E99p1ant. All rights reserved.

package dbutil

var DefaultPageSize = 20

type Pagination struct {
	Page     int
	PageSize int
}

// LimitOffset returns LIMIT and OFFSET parameter for SQL.
// The first page is page 0.
func LimitOffset(page, pageSize int) (limit, offset int) {
	if page <= 0 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	return pageSize, (page - 1) * pageSize
}
