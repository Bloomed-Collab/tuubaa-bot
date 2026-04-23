package level

type deque[T any] struct {
	items []T
}

func newDeque[T any]() *deque[T] {
	return &deque[T]{items: make([]T, 0)}
}

func (d *deque[T]) pushFront(item T) {
	d.items = append([]T{item}, d.items...)
}

func (d *deque[T]) popBack() (T, bool) {
	var zero T
	if len(d.items) == 0 {
		return zero, false
	}
	last := d.items[len(d.items)-1]
	d.items = d.items[:len(d.items)-1]
	return last, true
}

func (d *deque[T]) peekBack() (T, bool) {
	var zero T
	if len(d.items) == 0 {
		return zero, false
	}
	return d.items[len(d.items)-1], true
}

func (d *deque[T]) peekAt(index int) (T, bool) {
	var zero T
	if index < 0 || index >= len(d.items) {
		return zero, false
	}
	return d.items[index], true
}

func (d *deque[T]) size() int {
	return len(d.items)
}
