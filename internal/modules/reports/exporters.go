package reports

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type ExcelGenerator struct{}

func NewExcelGenerator() *ExcelGenerator {
	return &ExcelGenerator{}
}

func strPtr(s string) *string {
	return &s
}

func (g *ExcelGenerator) createStyles(f *excelize.File) (headerStyle, dataStyle, moneyStyle int) {
	headerStyle, _ = f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#000000"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	dataStyle, _ = f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	moneyStyle, _ = f.NewStyle(&excelize.Style{
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		CustomNumFmt: strPtr("#,##0.00"),
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	return headerStyle, dataStyle, moneyStyle
}

func (g *ExcelGenerator) GenerateProfit(data []ProfitItem) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Profit"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, dataStyle, moneyStyle := g.createStyles(f)

	headers := []string{"Артикул", "Товар", "Продано", "Выручка", "Себестоимость", "Валовая прибыль", "Налог", "Чистая прибыль", "Рентабельность %"}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	lastRow := len(data) + 1
	for i, item := range data {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.SKU)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.ProductName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.QuantitySold.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.SalesTotal.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.CostTotal.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item.GrossProfit.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item.EstimatedTax.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), item.NetProfit.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), item.Profitability.InexactFloat64())
	}

	f.SetCellStyle(sheet, "A1", "I1", headerStyle)

	if len(data) > 0 {
		f.SetCellStyle(sheet, "A2", fmt.Sprintf("C%d", lastRow), dataStyle)
		f.SetCellStyle(sheet, "D2", fmt.Sprintf("I%d", lastRow), moneyStyle)
	}

	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "I", 15)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *ExcelGenerator) GenerateStock(data []StockItem) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Stock"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, dataStyle, moneyStyle := g.createStyles(f)

	headers := []string{"Склад", "Категория", "Артикул", "Товар", "Ед. изм.", "Остаток", "Ср. цена", "Сумма"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	lastRow := len(data) + 1
	for i, item := range data {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.WarehouseName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.Category)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.SKU)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.ProductName)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.Unit)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item.Quantity.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item.AvgPurchasePrice.InexactFloat64())
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), item.TotalValue.InexactFloat64())
	}

	f.SetCellStyle(sheet, "A1", "H1", headerStyle)

	if len(data) > 0 {
		f.SetCellStyle(sheet, "A2", fmt.Sprintf("F%d", lastRow), dataStyle)
		f.SetCellStyle(sheet, "G2", fmt.Sprintf("H%d", lastRow), moneyStyle)
	}

	f.SetColWidth(sheet, "A", "B", 25)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 40)
	f.SetColWidth(sheet, "E", "F", 12)
	f.SetColWidth(sheet, "G", "H", 15)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *ExcelGenerator) GenerateMovements(data []MovementItem) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Movements"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, dataStyle, _ := g.createStyles(f)

	headers := []string{"Дата", "Документ", "Номер", "Склад", "Товар", "SKU", "Ед.", "Приход", "Расход"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	lastRow := len(data) + 1
	for i, item := range data {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.Date.Format("02.01.2006 15:04"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.DocumentType)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.DocumentNumber)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.WarehouseName)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.ProductName)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item.SKU)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), item.Unit)

		if item.QuantityIn.IsPositive() {
			f.SetCellValue(sheet, fmt.Sprintf("H%d", row), item.QuantityIn.InexactFloat64())
		}
		if item.QuantityOut.IsPositive() {
			f.SetCellValue(sheet, fmt.Sprintf("I%d", row), item.QuantityOut.InexactFloat64())
		}
	}

	f.SetCellStyle(sheet, "A1", "I1", headerStyle)
	if len(data) > 0 {
		f.SetCellStyle(sheet, "A2", fmt.Sprintf("I%d", lastRow), dataStyle)
	}

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "C", 15)
	f.SetColWidth(sheet, "D", "D", 25)
	f.SetColWidth(sheet, "E", "E", 40)
	f.SetColWidth(sheet, "F", "I", 12)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *ExcelGenerator) GenerateCustomers(data []CustomerReportItem) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Customers"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, dataStyle, moneyStyle := g.createStyles(f)

	headers := []string{"Контрагент", "Кол-во сделок", "Сумма покупок"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	lastRow := len(data) + 1
	for i, item := range data {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.CounterpartyName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.OperationsCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.TotalRevenue.InexactFloat64())
	}

	f.SetCellStyle(sheet, "A1", "C1", headerStyle)

	if len(data) > 0 {
		f.SetCellStyle(sheet, "A2", fmt.Sprintf("B%d", lastRow), dataStyle)
		f.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", lastRow), moneyStyle)
	}

	f.SetColWidth(sheet, "A", "A", 40)
	f.SetColWidth(sheet, "B", "B", 20)
	f.SetColWidth(sheet, "C", "C", 25)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type CSVGenerator struct{}

func NewCSVGenerator() *CSVGenerator { return &CSVGenerator{} }

func (g *CSVGenerator) writeBOM(buf *bytes.Buffer) {
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
}

func (g *CSVGenerator) GenerateProfit(data []ProfitItem) ([]byte, error) {
	buf := &bytes.Buffer{}
	g.writeBOM(buf)
	w := csv.NewWriter(buf)
	w.Comma = ';'
	w.Write([]string{"Артикул", "Товар", "Продано", "Выручка", "Себестоимость", "Валовая прибыль", "Налог", "Чистая прибыль", "Рентабельность %"})
	for _, item := range data {
		w.Write([]string{
			item.SKU, item.ProductName, item.QuantitySold.StringFixed(0),
			item.SalesTotal.StringFixed(2), item.CostTotal.StringFixed(2),
			item.GrossProfit.StringFixed(2), item.EstimatedTax.StringFixed(2),
			item.NetProfit.StringFixed(2), item.Profitability.StringFixed(1),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func (g *CSVGenerator) GenerateStock(data []StockItem) ([]byte, error) {
	buf := &bytes.Buffer{}
	g.writeBOM(buf)
	w := csv.NewWriter(buf)
	w.Comma = ';'
	w.Write([]string{"Склад", "Категория", "Артикул", "Товар", "Ед. изм.", "Остаток", "Ср. цена", "Сумма"})
	for _, item := range data {
		w.Write([]string{
			item.WarehouseName, item.Category, item.SKU, item.ProductName, item.Unit,
			item.Quantity.StringFixed(2), item.AvgPurchasePrice.StringFixed(2), item.TotalValue.StringFixed(2),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func (g *CSVGenerator) GenerateMovements(data []MovementItem) ([]byte, error) {
	buf := &bytes.Buffer{}
	g.writeBOM(buf)
	w := csv.NewWriter(buf)
	w.Comma = ';'
	w.Write([]string{"Дата", "Документ", "Номер", "Склад", "Товар", "SKU", "Ед.", "Приход", "Расход"})
	for _, item := range data {
		inStr := ""
		if item.QuantityIn.IsPositive() {
			inStr = item.QuantityIn.StringFixed(2)
		}
		outStr := ""
		if item.QuantityOut.IsPositive() {
			outStr = item.QuantityOut.StringFixed(2)
		}
		w.Write([]string{
			item.Date.Format("02.01.2006 15:04"), item.DocumentType, item.DocumentNumber,
			item.WarehouseName, item.ProductName, item.SKU, item.Unit, inStr, outStr,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func (g *CSVGenerator) GenerateCustomers(data []CustomerReportItem) ([]byte, error) {
	buf := &bytes.Buffer{}
	g.writeBOM(buf)
	w := csv.NewWriter(buf)
	w.Comma = ';'
	w.Write([]string{"Контрагент", "Кол-во сделок", "Сумма покупок"})
	for _, item := range data {
		w.Write([]string{
			item.CounterpartyName, fmt.Sprintf("%d", item.OperationsCount), item.TotalRevenue.StringFixed(2),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
