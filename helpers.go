package main

// paginate returns a single page of an interface slice.
func paginate(total int, page int) (offset, limit int) {
	if page < 1 {
		page = 1
	}
	offset = (page - 1) * pageSize
	if offset >= total {
		return offset, 0
	}
	limit = pageSize
	if offset+limit > total {
		limit = total - offset
	}
	return offset, limit
}
