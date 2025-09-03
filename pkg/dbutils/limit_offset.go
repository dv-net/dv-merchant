package dbutils

import (
	"fmt"

	"github.com/huandu/go-sqlbuilder"
)

func BuildLimit(b sqlbuilder.Builder, limit int) sqlbuilder.Builder {
	return sqlbuilder.Build("$0 LIMIT $1", b, sqlbuilder.Raw(fmt.Sprintf("%d", limit)))
}

func BuildOffset(b sqlbuilder.Builder, offset int) sqlbuilder.Builder {
	return sqlbuilder.Build("$0 OFFSET $1", b, sqlbuilder.Raw(fmt.Sprintf("%d", offset)))
}
