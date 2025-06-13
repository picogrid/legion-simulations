package models

type SortField struct {
	Field     string
	Ascending bool
}

// Paging contains optional paging information.
// @Description Contains optional paging information like next and previous page numbers.
// @name Paging
type Paging struct {
	// Next is the next page number.
	// @description Next is the next page number.
	Next *int `json:"next" swaggertype:"integer" example:"2"`
	// Previous is the previous page number.
	// @description Previous is the previous page number.
	Previous *int `json:"previous" swaggertype:"integer" example:"0"`
}

// PaginatedResponse that leverages a slice of type T.
// @Description PaginatedResponse is a generic struct that contains a slice of results of type T, total count of items, and optional paging information.
// @name PaginatedResponse
type PaginatedResponse[T any] struct {
	// Results is a slice of items of type T.
	// @description Results is a slice of items of type T.
	Results []T `json:"results"`
	// TotalCount is the total number of items available.
	// @description TotalCount is the total number of items available.
	TotalCount int `json:"total_count" swaggertype:"integer" example:"100"`
	// Paging contains optional paging information.
	// @description Paging contains optional paging information.
	Paging Paging `json:"paging,omitempty" swaggertype:"object"`
}

// NewPaginatedResponse creates a new PaginatedResponse for a given slice of type T.
func NewPaginatedResponse[T any](results []T, totalCount int, next *int, previous *int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Results:    results,
		TotalCount: totalCount,
		Paging:     Paging{Next: next, Previous: previous},
	}
}

func NewEmptyPaginatedResponse[T any]() PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Results:    []T{},
		TotalCount: 0,
		Paging:     Paging{},
	}
}
