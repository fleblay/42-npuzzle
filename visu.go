package main

import (
	"log"

	"strconv"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// Visu is the main function for the visualisation
func convertBoard(board [][]int) [][]string {
	var convertedBoard [][]string
	for i := 0; i < len(board); i++ {
		var row []string
		for j := 0; j < len(board); j++ {
			if board[i][j] == 0 {
				row = append(row, " ")
				continue
			}
			row = append(row, strconv.Itoa(board[i][j]))
		}
		convertedBoard = append(convertedBoard, row)
	}
	return convertedBoard

}

func PrintBoard(board [][]int) bool {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	table := createTable(board)

	ui.Render(table)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return false
		case "s":
			moveUp(board, getEmptySpot(board))
		case "w":
			moveDown(board, getEmptySpot(board))
		case "d":
			moveLeft(board, getEmptySpot(board))
		case "a":
			moveRight(board, getEmptySpot(board))
		}
		if isEqual(board, goal(len(board))) {
			return handleWinScenario()
		} else {
			table.Rows = convertBoard(board)
			ui.Render(table)
		}

		updateTableRows(table, board)
		ui.Render(table)
	}
}

func createTable(board [][]int) *widgets.Table {
	table := widgets.NewTable()
	table.Title = "n-puzzle"
	table.TitleStyle = ui.NewStyle(ui.ColorBlue)
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.BorderStyle = ui.NewStyle(ui.ColorGreen)
	table.SetRect(0, 0, len(board)*6, len(board)*2+1)
	table.FillRow = true
	table.TextAlignment = ui.AlignCenter
	updateTableRows(table, board)
	return table
}

func updateTableRows(table *widgets.Table, board [][]int) {
	table.Rows = convertBoard(board)
}

func handleWinScenario() bool {
	ui.Clear()
	p := createWinParagraph()
	ui.Render(p)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "n", "<C-c>":
			return false
		case "y":
			return true
		}
	}
}

func createWinParagraph() *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Text = "You won! Do you want to restart? (y/n)"
	p.SetRect(0, 0, 25, 5)
	p.TextStyle = ui.NewStyle(ui.ColorGreen)
	p.BorderStyle = ui.NewStyle(ui.ColorGreen)
	return p
}
