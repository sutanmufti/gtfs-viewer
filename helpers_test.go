package main

import "testing"

func TestPaginate(t *testing.T) {
	tests := []struct {
		name           string
		total          int
		page           int
		wantOffset     int
		wantLimit      int
		wantTotalPages int
	}{
		{"zero total", 0, 1, 0, 0, 1},
		{"first page full", 25, 1, 0, 10, 3},
		{"second page full", 25, 2, 10, 10, 3},
		{"last page partial", 25, 3, 20, 5, 3},
		{"page beyond total", 25, 99, 980, 0, 3},
		{"page zero clamped to 1", 10, 0, 0, 10, 1},
		{"negative page clamped to 1", 10, -5, 0, 10, 1},
		{"exact multiple of pageSize", 20, 2, 10, 10, 2},
		{"single record", 1, 1, 0, 1, 1},
		{"exactly pageSize records", 10, 1, 0, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, limit, totalPages := paginate(tt.total, tt.page)
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
			if totalPages != tt.wantTotalPages {
				t.Errorf("totalPages = %d, want %d", totalPages, tt.wantTotalPages)
			}
		})
	}
}
