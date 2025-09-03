package dbutils

import "fmt"

type PaginationOption func(*paginationOptions)

type paginationOptions struct {
	MaxLimit uint32
}

func WithMaxLimit(v uint32) PaginationOption {
	return func(o *paginationOptions) {
		o.MaxLimit = v
	}
}

func Pagination(page, pageSize *uint32, opts ...PaginationOption) (limit, offset uint32, err error) {
	options := paginationOptions{
		MaxLimit: 100,
	}

	for _, opt := range opts {
		opt(&options)
	}

	limit = 100
	offset = 0

	if pageSize != nil {
		limit = *pageSize
	}

	if page != nil && *page != 0 {
		offset = (*page - 1) * limit
	}

	if limit < 1 {
		return 0, 0, fmt.Errorf("pagination error: page size cannot be less than 1")
	}

	if limit > options.MaxLimit {
		return 0, 0, fmt.Errorf("pagination error: page size cannot be greater than %d", options.MaxLimit)
	}

	return
}
