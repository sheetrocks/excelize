// Copyright 2016 - 2022 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to and
// read from XLAM / XLSM / XLSX / XLTM / XLTX files. Supports reading and
// writing spreadsheet documents generated by Microsoft Excel™ 2007 and later.
// Supports complex components by high compatibility, and provided streaming
// API for generating or reading data from a worksheet with huge amounts of
// data. This library needs Go version 1.15 or later.

package excelize

import "strings"

// Rect gets merged cell rectangle coordinates sequence.
func (mc *xlsxMergeCell) Rect() ([]int, error) {
	var err error
	if mc.rect == nil {
		mc.rect, err = areaRefToCoordinates(mc.Ref)
	}
	return mc.rect, err
}

// MergeCell provides a function to merge cells by given coordinate area and
// sheet name. Merging cells only keeps the upper-left cell value, and
// discards the other values. For example create a merged cell of D3:E9 on
// Sheet1:
//
//	err := f.MergeCell("Sheet1", "D3", "E9")
//
// If you create a merged cell that overlaps with another existing merged cell,
// those merged cells that already exist will be removed. The cell coordinates
// tuple after merging in the following range will be: A1(x3,y1) D1(x2,y1)
// A8(x3,y4) D8(x2,y4)
//
//	             B1(x1,y1)      D1(x2,y1)
//	           +------------------------+
//	           |                        |
//	 A4(x3,y3) |    C4(x4,y3)           |
//	+------------------------+          |
//	|          |             |          |
//	|          |B5(x1,y2)    | D5(x2,y2)|
//	|          +------------------------+
//	|                        |
//	|A8(x3,y4)      C8(x4,y4)|
//	+------------------------+
func (f *File) MergeCell(sheet, hCell, vCell string) error {
	rect, err := areaRefToCoordinates(hCell + ":" + vCell)
	if err != nil {
		return err
	}
	// Correct the coordinate area, such correct C1:B3 to B1:C3.
	_ = sortCoordinates(rect)

	hCell, _ = CoordinatesToCellName(rect[0], rect[1])
	vCell, _ = CoordinatesToCellName(rect[2], rect[3])

	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	ref := hCell + ":" + vCell
	if ws.MergeCells != nil {
		ws.MergeCells.Cells = append(ws.MergeCells.Cells, &xlsxMergeCell{Ref: ref, rect: rect})
	} else {
		ws.MergeCells = &xlsxMergeCells{Cells: []*xlsxMergeCell{{Ref: ref, rect: rect}}}
	}
	ws.MergeCells.Count = len(ws.MergeCells.Cells)
	return err
}

// UnmergeCell provides a function to unmerge a given coordinate area.
// For example unmerge area D3:E9 on Sheet1:
//
//	err := f.UnmergeCell("Sheet1", "D3", "E9")
//
// Attention: overlapped areas will also be unmerged.
func (f *File) UnmergeCell(sheet string, hCell, vCell string) error {
	ws, err := f.SRworkSheetReader(sheet)
	if err != nil {
		return err
	}
	rect1, err := areaRefToCoordinates(hCell + ":" + vCell)
	if err != nil {
		return err
	}

	// Correct the coordinate area, such correct C1:B3 to B1:C3.
	_ = sortCoordinates(rect1)

	// return nil since no MergeCells in the sheet
	if ws.MergeCells == nil {
		return nil
	}
	if err = f.mergeOverlapCells(ws); err != nil {
		return err
	}
	i := 0
	for _, mergeCell := range ws.MergeCells.Cells {
		if mergeCell == nil {
			continue
		}
		rect2, _ := areaRefToCoordinates(mergeCell.Ref)
		if isOverlap(rect1, rect2) {
			continue
		}
		ws.MergeCells.Cells[i] = mergeCell
		i++
	}
	ws.MergeCells.Cells = ws.MergeCells.Cells[:i]
	ws.MergeCells.Count = len(ws.MergeCells.Cells)
	if ws.MergeCells.Count == 0 {
		ws.MergeCells = nil
	}
	return nil
}

// GetMergeCells provides a function to get all merged cells from a worksheet
// currently.
func (f *File) GetMergeCells(sheet string) ([]MergeCell, error) {
	var mergeCells []MergeCell
	ws, err := f.SRworkSheetReader(sheet)
	if err != nil {
		return mergeCells, err
	}
	if ws.MergeCells != nil {
		if err = f.SRmergeOverlapCells(ws); err != nil {
			return mergeCells, err
		}
		mergeCells = make([]MergeCell, 0, len(ws.MergeCells.Cells))
		for i := range ws.MergeCells.Cells {
			ref := ws.MergeCells.Cells[i].Ref
			axis := strings.Split(ref, ":")[0]
			val, _ := f.SRGetCellValue(sheet, axis)
			mergeCells = append(mergeCells, []string{ref, val})
		}
	}
	return mergeCells, err
}

// overlapRange calculate overlap range of merged cells, and returns max
// column and rows of the range.
func overlapRange(ws *SRxlsxWorksheet) (row, col int, err error) {
	var rect []int
	for _, mergeCell := range ws.MergeCells.Cells {
		if mergeCell == nil {
			continue
		}
		if rect, err = mergeCell.Rect(); err != nil {
			return
		}
		x1, y1, x2, y2 := rect[0], rect[1], rect[2], rect[3]
		if x1 > col {
			col = x1
		}
		if x2 > col {
			col = x2
		}
		if y1 > row {
			row = y1
		}
		if y2 > row {
			row = y2
		}
	}
	return
}

// overlapRange calculate overlap range of merged cells, and returns max
// column and rows of the range.
func SRoverlapRange(ws *SRxlsxWorksheet) (row, col int, err error) {
	var rect []int
	for _, mergeCell := range ws.MergeCells.Cells {
		if mergeCell == nil {
			continue
		}
		if rect, err = mergeCell.Rect(); err != nil {
			return
		}
		x1, y1, x2, y2 := rect[0], rect[1], rect[2], rect[3]
		if x1 > col {
			col = x1
		}
		if x2 > col {
			col = x2
		}
		if y1 > row {
			row = y1
		}
		if y2 > row {
			row = y2
		}
	}
	return
}

// flatMergedCells convert merged cells range reference to cell-matrix.
func flatMergedCells(ws *SRxlsxWorksheet, matrix [][]*xlsxMergeCell) error {
	for i, cell := range ws.MergeCells.Cells {
		rect, err := cell.Rect()
		if err != nil {
			return err
		}
		x1, y1, x2, y2 := rect[0]-1, rect[1]-1, rect[2]-1, rect[3]-1
		var overlapCells []*xlsxMergeCell
		for x := x1; x <= x2; x++ {
			for y := y1; y <= y2; y++ {
				if matrix[x][y] != nil {
					overlapCells = append(overlapCells, matrix[x][y])
				}
				matrix[x][y] = cell
			}
		}
		if len(overlapCells) != 0 {
			newCell := cell
			for _, overlapCell := range overlapCells {
				newCell = mergeCell(cell, overlapCell)
			}
			newRect, _ := newCell.Rect()
			x1, y1, x2, y2 := newRect[0]-1, newRect[1]-1, newRect[2]-1, newRect[3]-1
			for x := x1; x <= x2; x++ {
				for y := y1; y <= y2; y++ {
					matrix[x][y] = newCell
				}
			}
			ws.MergeCells.Cells[i] = newCell
		}
	}
	return nil
}

// flatMergedCells convert merged cells range reference to cell-matrix.
func SRflatMergedCells(ws *SRxlsxWorksheet, matrix [][]*xlsxMergeCell) error {
	for i, cell := range ws.MergeCells.Cells {
		rect, err := cell.Rect()
		if err != nil {
			return err
		}
		x1, y1, x2, y2 := rect[0]-1, rect[1]-1, rect[2]-1, rect[3]-1
		var overlapCells []*xlsxMergeCell
		for x := x1; x <= x2; x++ {
			for y := y1; y <= y2; y++ {
				if matrix[x][y] != nil {
					overlapCells = append(overlapCells, matrix[x][y])
				}
				matrix[x][y] = cell
			}
		}
		if len(overlapCells) != 0 {
			newCell := cell
			for _, overlapCell := range overlapCells {
				newCell = mergeCell(cell, overlapCell)
			}
			newRect, _ := newCell.Rect()
			x1, y1, x2, y2 := newRect[0]-1, newRect[1]-1, newRect[2]-1, newRect[3]-1
			for x := x1; x <= x2; x++ {
				for y := y1; y <= y2; y++ {
					matrix[x][y] = newCell
				}
			}
			ws.MergeCells.Cells[i] = newCell
		}
	}
	return nil
}

// mergeOverlapCells merge overlap cells.
func (f *File) mergeOverlapCells(ws *SRxlsxWorksheet) error {
	rows, cols, err := overlapRange(ws)
	if err != nil {
		return err
	}
	if rows == 0 || cols == 0 {
		return nil
	}
	matrix := make([][]*xlsxMergeCell, cols)
	for i := range matrix {
		matrix[i] = make([]*xlsxMergeCell, rows)
	}
	_ = flatMergedCells(ws, matrix)
	mergeCells := ws.MergeCells.Cells[:0]
	for _, cell := range ws.MergeCells.Cells {
		rect, _ := cell.Rect()
		x1, y1, x2, y2 := rect[0]-1, rect[1]-1, rect[2]-1, rect[3]-1
		if matrix[x1][y1] == cell {
			mergeCells = append(mergeCells, cell)
			for x := x1; x <= x2; x++ {
				for y := y1; y <= y2; y++ {
					matrix[x][y] = nil
				}
			}
		}
	}
	ws.MergeCells.Count, ws.MergeCells.Cells = len(mergeCells), mergeCells
	return nil
}

// mergeOverlapCells merge overlap cells.
func (f *File) SRmergeOverlapCells(ws *SRxlsxWorksheet) error {
	rows, cols, err := SRoverlapRange(ws)
	if err != nil {
		return err
	}
	if rows == 0 || cols == 0 {
		return nil
	}
	matrix := make([][]*xlsxMergeCell, cols)
	for i := range matrix {
		matrix[i] = make([]*xlsxMergeCell, rows)
	}
	_ = SRflatMergedCells(ws, matrix)
	mergeCells := ws.MergeCells.Cells[:0]
	for _, cell := range ws.MergeCells.Cells {
		rect, _ := cell.Rect()
		x1, y1, x2, y2 := rect[0]-1, rect[1]-1, rect[2]-1, rect[3]-1
		if matrix[x1][y1] == cell {
			mergeCells = append(mergeCells, cell)
			for x := x1; x <= x2; x++ {
				for y := y1; y <= y2; y++ {
					matrix[x][y] = nil
				}
			}
		}
	}
	ws.MergeCells.Count, ws.MergeCells.Cells = len(mergeCells), mergeCells
	return nil
}

// mergeCell merge two cells.
func mergeCell(cell1, cell2 *xlsxMergeCell) *xlsxMergeCell {
	rect1, _ := cell1.Rect()
	rect2, _ := cell2.Rect()

	if rect1[0] > rect2[0] {
		rect1[0], rect2[0] = rect2[0], rect1[0]
	}

	if rect1[2] < rect2[2] {
		rect1[2], rect2[2] = rect2[2], rect1[2]
	}

	if rect1[1] > rect2[1] {
		rect1[1], rect2[1] = rect2[1], rect1[1]
	}

	if rect1[3] < rect2[3] {
		rect1[3], rect2[3] = rect2[3], rect1[3]
	}
	hCell, _ := CoordinatesToCellName(rect1[0], rect1[1])
	vCell, _ := CoordinatesToCellName(rect1[2], rect1[3])
	return &xlsxMergeCell{rect: rect1, Ref: hCell + ":" + vCell}
}

// MergeCell define a merged cell data.
// It consists of the following structure.
// example: []string{"D4:E10", "cell value"}
type MergeCell []string

// GetCellValue returns merged cell value.
func (m *MergeCell) GetCellValue() string {
	return (*m)[1]
}

// GetStartAxis returns the top left cell coordinates of merged range, for
// example: "C2".
func (m *MergeCell) GetStartAxis() string {
	axis := strings.Split((*m)[0], ":")
	return axis[0]
}

// GetEndAxis returns the bottom right cell coordinates of merged range, for
// example: "D4".
func (m *MergeCell) GetEndAxis() string {
	axis := strings.Split((*m)[0], ":")
	return axis[1]
}
