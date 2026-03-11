package main

import gtfsparser "github.com/sutanmufti/gtfs-parser"

// loadGTFS parses and compiles a GTFS zip file from the given path.
func loadGTFS(path string) (*gtfsparser.GTFS, error) {
	g := &gtfsparser.GTFS{FileName: path}
	if err := g.ParseAll(); err != nil {
		return nil, err
	}
	g.Compile()
	return g, nil
}

// paginate returns the offset, limit, and total page count for a given page.
func paginate(total int, page int) (offset, limit, totalPages int) {
	if page < 1 {
		page = 1
	}
	totalPages = (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	offset = (page - 1) * pageSize
	if offset >= total {
		return offset, 0, totalPages
	}
	limit = pageSize
	if offset+limit > total {
		limit = total - offset
	}
	return offset, limit, totalPages
}
