package service

import "github.com/officeryoda/dozingo/internal/generated"

func countCompleteBingos(cells []generated.GameCell, size int32) int32 {
	if size == 0 || len(cells) != int(size*size) {
		return 0
	}

	// Build a flat boolean grid indexed by position.
	marked := make([]bool, size*size)
	for i := range cells {
		c := &cells[i]
		if c.Position >= 0 && c.Position < size*size {
			marked[c.Position] = c.IsMarked
		}
	}

	var count int32

	// Check each row.
	for row := range size {
		if rowComplete(marked, row, size) {
			count++
		}
	}

	// Check each column.
	for col := range size {
		if colComplete(marked, col, size) {
			count++
		}
	}

	// Main diagonal: top-left → bottom-right .
	if diagComplete(marked, size) {
		count++
	}

	// Anti-diagonal: top-right → bottom-left .
	if antiDiagComplete(marked, size) {
		count++
	}

	return count
}

func rowComplete(marked []bool, row, size int32) bool {
	start := row * size
	for col := range size {
		if !marked[start+col] {
			return false
		}
	}

	return true
}

func colComplete(marked []bool, col, size int32) bool {
	for row := range size {
		if !marked[row*size+col] {
			return false
		}
	}

	return true
}

func diagComplete(marked []bool, size int32) bool {
	for i := range size {
		if !marked[i*size+i] {
			return false
		}
	}

	return true
}

func antiDiagComplete(marked []bool, size int32) bool {
	for i := range size {
		if !marked[i*size+(size-1-i)] {
			return false
		}
	}

	return true
}
