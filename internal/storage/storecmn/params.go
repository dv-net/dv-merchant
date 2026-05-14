package storecmn

import "fmt"

// SafeOrderBy validates the column against an allowlist and returns the safe value.
// Returns an error if the column is not in the allowlist.
func SafeOrderBy(column string, allowlist map[string]string) (string, error) {
	if column == "" {
		return "", nil
	}
	safe, ok := allowlist[column]
	if !ok {
		return "", fmt.Errorf("invalid sort column: %q", column)
	}
	return safe, nil
}

type CommonFindParams struct {
	IsAscOrdering bool   `json:"is_asc_ordering"`
	OrderBy       string `json:"order_by"`
	PageParams
} //	@name	CommonFindParams

func NewCommonFindParams() *CommonFindParams {
	return &CommonFindParams{}
}

func (s *CommonFindParams) SetIsAscOrdering(v bool) *CommonFindParams {
	s.IsAscOrdering = v
	return s
}

func (s *CommonFindParams) SetOrderBy(v string) *CommonFindParams {
	s.OrderBy = v
	return s
}

func (s *CommonFindParams) SetPage(v *uint32) *CommonFindParams {
	s.Page = v
	return s
}

func (s *CommonFindParams) SetPageSize(v *uint32) *CommonFindParams {
	s.PageSize = v
	return s
}

type PageParams struct {
	Page     *uint32
	PageSize *uint32
} //	@name	PageParams

type FindResponseWithPagingFlag[T any] struct {
	Items            []T  `json:"items"`
	IsNextPageExists bool `json:"is_next_page_exists"`
} //	@name	ResponseWithPagingFlag

type FullPagingData struct {
	Total    uint64 `json:"total"`
	PageSize uint64 `json:"page_size"`
	Page     uint64 `json:"page"`
	LastPage uint64 `json:"last_page"`
} //	@name	FullPagingData

type FindResponseWithFullPagination[T any] struct {
	Items      []T            `json:"items"`
	Pagination FullPagingData `json:"pagination"`
} //	@name	ResponseWithFullPagination
