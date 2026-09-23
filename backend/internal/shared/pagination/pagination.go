package pagination

const (
	defaultPageSize = 20
	firstPage       = 1
)

type Query struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

type Page[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

func (q Query) Number() int {
	if q.Page < firstPage {
		return firstPage
	}
	return q.Page
}

func (q Query) Size() int {
	if q.PageSize <= 0 {
		return defaultPageSize
	}
	return q.PageSize
}

func (q Query) Offset() int {
	return (q.Number() - 1) * q.Size()
}

func New[T any](items []T, query Query, total int64) Page[T] {
	size := query.Size()
	return Page[T]{
		Items:      items,
		Page:       query.Number(),
		PageSize:   size,
		TotalItems: total,
		TotalPages: int((total + int64(size) - 1) / int64(size)),
	}
}

func Map[S, T any](items []S, convert func(S) T) []T {
	converted := make([]T, 0, len(items))
	for _, item := range items {
		converted = append(converted, convert(item))
	}
	return converted
}
