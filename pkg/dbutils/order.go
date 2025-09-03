package dbutils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dv-net/mx/util"
	"github.com/huandu/go-sqlbuilder"
)

var ErrOrderByIsNotValid = errors.New("order_by field is not valid")

type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "asc"
	OrderDirectionDesc OrderDirection = "desc"
)

func (s OrderDirection) Valid() bool {
	switch s {
	case OrderDirectionAsc, OrderDirectionDesc:
		return true
	default:
		return false
	}
}

type OrderBy[T any] struct {
	Attribute T
	Direction OrderDirection
}

type ValidableString interface {
	~string
	Validate() error
}

type StringOrderBy[T ValidableString] string

func (s StringOrderBy[T]) Validate() error {
	if len(s) > 0 {
		orderByStrings := strings.Split(string(s), ",")
		for _, orderByString := range orderByStrings {
			strs := strings.Split(orderByString, " ")
			if len(strs) != 2 {
				return ErrOrderByIsNotValid
			}
			if err := T(strs[0]).Validate(); err != nil {
				return ErrOrderByIsNotValid
			}
			if !OrderDirection(strs[1]).Valid() {
				return ErrOrderByIsNotValid
			}
		}
	}
	return nil
}

func (s StringOrderBy[T]) ToStruct() ([]*OrderBy[T], error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	orderBy := []*OrderBy[T]{}
	if len(s) > 0 {
		orderByStrings := strings.Split(string(s), ",")
		for _, orderByString := range orderByStrings {
			orderByStrings := strings.Split(orderByString, " ")

			orderBy = append(orderBy, util.Pointer(OrderBy[T]{
				Attribute: T(orderByStrings[0]),
				Direction: OrderDirection(orderByStrings[1]),
			}))
		}
	}

	return orderBy, nil
}

func BuildOrderBy[T ValidableString](b sqlbuilder.Builder, orderBy StringOrderBy[T]) (sqlbuilder.Builder, error) {
	orderByStruct, err := orderBy.ToStruct()
	if err != nil {
		return nil, err
	}

	orderByQuery := ""
	for _, orderBy := range orderByStruct {
		if orderBy != nil {
			orderByQuery += fmt.Sprintf("%s %s, ", orderBy.Attribute, orderBy.Direction)
		}
	}
	orderByQuery = strings.TrimSuffix(orderByQuery, ", ")

	if len(orderByQuery) == 0 {
		return b, nil
	}

	return sqlbuilder.Build("$0 ORDER BY $1", b, sqlbuilder.Raw(orderByQuery)), nil
}
