package mapgen

import (
	"encoding/json"
)

type CellType bool

const (
	EmptySpaceCell CellType = false
	WallCell       CellType = true
)

func (c CellType) MarshalJSON() ([]byte, error) {
	n := 0
	if c {
		n = 1
	}
	return json.Marshal(n)
}

type Map struct {
	Width  int
	Height int
	Cells  [][]CellType
}

func NewMap(w int, h int) *Map {
	m := &Map{
		Width:  w,
		Height: h,
		// Cells:  empty grid generated later,
	}

	// Generating the grid
	{
		// Allocate the top-level slice.
		cells := make([][]CellType, w) // One row per unit of x.
		// Loop over the rows, allocating the z-slice for each x.
		for x := range cells {
			cells[x] = make([]CellType, h)
		}

		m.Cells = cells
	}

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			if (i == 0) || (j == 0) || (i == w-1) || (j == h-1) {
				m.Cells[i][j] = WallCell
			} else if (i%4 == 0) && (j%4 == 0) {
				m.Cells[i][j] = WallCell
			}
		}
	}

	return m
}
