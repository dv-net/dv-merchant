package search_response

type SearchByCriteriaResponse[T any] struct {
	SearchType string `json:"search_type"`
	Data       T      `json:"data"`
} //	@name	SearchByCriteriaResponse

func PrepareByCriteria[T any](searchType string, data T) SearchByCriteriaResponse[T] {
	return SearchByCriteriaResponse[T]{
		SearchType: searchType,
		Data:       data,
	}
}
