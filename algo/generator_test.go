package algo

import (
	"testing"
)

func TestGridGenerator(t *testing.T) {
	test := []int{3, 4, 5, 6, 7, 8}
	for _, test := range test {
		values := map[int]int{}
		grid := GridGenerator(test)
		for _, row := range grid {
			for _, item := range row {
				if _, ok := values[item]; ok {
					t.Errorf("Duplicate key %d", item)
				} else {
					values[item] = item
				}
			}
		}
	}
}

func TestGoal(t *testing.T) {

	test := []struct {
		mapSize int
		goal    [][]int
	}{
		{3, [][]int{{1, 2, 3}, {8, 0, 4}, {7, 6, 5}}},
		{4, [][]int{{1, 2, 3, 4}, {12, 13, 14, 5}, {11, 0, 15, 6}, {10, 9, 8, 7}}},
		{5, [][]int{{1, 2, 3, 4, 5}, {16, 17, 18, 19, 6}, {15, 24, 0, 20, 7}, {14, 23, 22, 21, 8}, {13, 12, 11, 10, 9}}},
	}
	for _, test := range test {
		if goal := goal(test.mapSize); isEqual(goal, test.goal) != true {
			t.Errorf("goal(%v) = %v", test.mapSize, goal)
		}
	}
}

func TestIsEqual(t *testing.T) {
	test := []struct {
		a, b [][]int
		want bool
	}{
		{[][]int{{1, 2, 3}, {8, 0, 4}, {7, 6, 5}}, [][]int{{1, 2, 3}, {8, 0, 4}, {7, 6, 5}}, true},
		{[][]int{{1, 2, 3}, {8, 0, 4}, {7, 5, 6}}, [][]int{{1, 2, 3}, {8, 0, 4}, {7, 6, 5}}, false},
	}

	for _, test := range test {
		if got := isEqual(test.a, test.b); got != test.want {
			t.Errorf("isEqual(%v, %v) = %v", test.a, test.b, got)
		}
	}

}

func TestmatrixToTableSnail(t *testing.T) {
	test := []struct {
		matrix [][]int
		want   []int
	}{
		{[][]int{{1, 2, 3}, {8, 0, 4}, {7, 6, 5}}, []int{1, 2, 3, 4, 5, 6, 7, 8, 0}},
		{[][]int{{1, 2, 3, 4}, {12, 13, 14, 5}, {11, 0, 15, 6}, {10, 9, 8, 7}}, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0}},
		{[][]int{{1, 2, 3, 4, 5}, {16, 17, 18, 19, 6}, {15, 24, 0, 20, 7}, {14, 23, 22, 21, 8}, {13, 12, 11, 10, 9}}, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 0}},
	}
	for _, test := range test {
		if got := matrixToTableSnail(test.matrix); isEqualTable(got, test.want) != true {
			t.Errorf("matrixToTableSnail(%v) = %v", test.matrix, got)
		}
	}

}

func TestIsSolvable(t *testing.T) {
	dir := "maps/solvables/"
	files := openDir(dir)
	for _, file := range files {
		openFile := dir + file.Name()
		_, matrix := ParseInput(OpenFile(openFile))
		if file.Name()[0:1] == "u" {
			if isSolvable(matrix) != false {
				t.Errorf("file [%v] isSolvable(%v) = %v", file.Name(), matrix, isSolvable(matrix))
			}
		} else {
			if isSolvable(matrix) != true {
				t.Errorf("file [%v] isSolvable(%v) = %v", file.Name(), matrix, isSolvable(matrix))
			}
		}
	}

}

func TestMatrixToString(t *testing.T) {

	test := []struct {
		matrix [][]int
		want   string
	}{
		{[][]int{{1, 2, 3}, {8, 0, 4}, {7, 6, 5}}, "1.2.3.8.0.4.7.6.5."},
		{[][]int{{1, 2, 3, 4}, {12, 13, 14, 5}, {11, 0, 15, 6}, {10, 9, 8, 7}}, "1.2.3.4.12.13.14.5.11.0.15.6.10.9.8.7."},
	}
	for _, test := range test {
		if got, _, _ := matrixToStringNoOpti(test.matrix, 1, 1); got != test.want {
			t.Errorf("matrixTo(%v) = %v", test.matrix, got)
		}
	}

}
