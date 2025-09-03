package dbutils

import "slices"

type FieldMask[T ~string] []T

func (s *FieldMask[T]) Add(items ...T) []T {
	if len(items) == 0 {
		return *s
	}

	for _, item := range items {
		if !slices.Contains(*s, item) {
			*s = append(*s, item)
		}
	}

	return *s
}

func (s FieldMask[T]) Items() []T {
	if s == nil {
		return FieldMask[T]{}
	}
	return s
}

func (s FieldMask[T]) Len() int {
	return len(s)
}

func (s FieldMask[T]) Contains(v T) bool {
	return slices.Contains(s, v)
}

func FieldMaskFromStrings[T ~string](s []string) FieldMask[T] {
	res := make([]T, len(s))
	for i, v := range s {
		res[i] = T(v)
	}
	return res
}
