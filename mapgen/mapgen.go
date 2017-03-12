package mapgen

import (
	"encoding/json"
	// "log"
	"math/rand"
)

type CellType bool

const (
	EmptySpaceCell CellType = true
	WallCell       CellType = false
)

func (c CellType) MarshalJSON() ([]byte, error) {
	n := 0
	if c == WallCell {
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
		Width:  (w+1)/2 - 1,
		Height: (h+1)/2 - 1,
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

	// First empty out the entire space, except the walls
	// for i := 0; i < w; i++ {
	// 	for j := 0; j < h; j++ {
	// 		if (i != 0) && (j != 0) && (i != w-1) && (j != h-1) {
	// 			m.Cells[i][j] = EmptySpaceCell
	// 		}
	// 	}
	// }

	// Pit a hole in the corner
	for x := 1; x < 5; x++ {
		for y := 1; y < 5; y++ {
			m.Cells[x][y] = EmptySpaceCell
		}
	}

	midX := (w + 1) / 2
	midY := (h + 1) / 2
	expandSize := 2
	for x := midX - expandSize; x <= midX+expandSize; x++ {
		for y := midY - expandSize; y <= midY+expandSize; y++ {
			m.Cells[x][y] = EmptySpaceCell
		}
	}

	// r for row、c for column
	// Generate random r
	r := rand.Intn(h)
	for r%2 == 0 {
		r = rand.Intn(h)
	}
	// Generate random c
	c := rand.Intn(w)
	for c%2 == 0 {
		c = rand.Intn(w)
	}
	// Starting cell
	m.Cells[r][c] = EmptySpaceCell

	//　Allocate the maze with recursive method
	recursion(m, r, c)

	return expandMap(m)
}

func expandMap(mini *Map) *Map {
	// m := &Map{
	// 	Width:  mini.Width * 2,
	// 	Height: mini.Height * 2,
	// }

	// // Allocate the top-level slice.
	// cells := make([][]CellType, m.Width) // One row per unit of x.
	// // Loop over the rows, allocating the z-slice for each x.
	// for x := range cells {
	// 	cells[x] = make([]CellType, m.Height)
	// }

	// m.Cells = cells

	return mini
}

func generateRandomDirections() []int {
	directions := []int{1, 2, 3, 4}
	for i := (len(directions) - 1); i > 0; i-- {
		j := rand.Intn(i)
		directions[j], directions[i] = directions[i], directions[j]
	}

	return directions
}

func recursion(m *Map, r int, c int) {

	maze := m.Cells

	// 4 random direction
	randDirs := generateRandomDirections()
	// Examine each direction
	for i := 0; i < len(randDirs); i++ {

		switch randDirs[i] {
		case 1: // Up
			//　Whether 2 cells up is out or not
			if r-2 <= 0 {
				continue
			}
			if maze[r-2][c] != EmptySpaceCell {
				maze[r-2][c] = EmptySpaceCell
				maze[r-1][c] = EmptySpaceCell
				recursion(m, r-2, c)
			}
			break
		case 2: // Right
			// Whether 2 cells to the right is out or not
			if c+2 >= m.Width-1 {
				continue
			}
			if maze[r][c+2] != EmptySpaceCell {
				maze[r][c+2] = EmptySpaceCell
				maze[r][c+1] = EmptySpaceCell
				recursion(m, r, c+2)
			}
			break
		case 3: // Down
			// Whether 2 cells down is out or not
			if r+2 >= m.Height-1 {
				continue
			}
			if maze[r+2][c] != EmptySpaceCell {
				maze[r+2][c] = EmptySpaceCell
				maze[r+1][c] = EmptySpaceCell
				recursion(m, r+2, c)
			}
			break
		case 4: // Left
			// Whether 2 cells to the left is out or not
			if c-2 <= 0 {
				continue
			}
			if maze[r][c-2] != EmptySpaceCell {
				maze[r][c-2] = EmptySpaceCell
				maze[r][c-1] = EmptySpaceCell
				recursion(m, r, c-2)
			}
			break
		}
	}

}
