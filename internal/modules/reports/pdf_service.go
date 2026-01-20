package reports

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/shopspring/decimal"
)

var (
	ColorBlack = []int{0, 0, 0}
	ColorGrey  = []int{240, 240, 240}
)

type PDFGenerator struct {
	fontPath string
}

func NewPDFGenerator(fontPath string) *PDFGenerator {
	return &PDFGenerator{fontPath: fontPath}
}

func (g *PDFGenerator) initPDF(orientation string) *fpdf.Fpdf {
	pdf := fpdf.New(orientation, "mm", "A4", "")
	pdf.SetFontLocation(g.fontPath)
	pdf.AddUTF8Font("Roboto", "", "Roboto-Regular.ttf")
	pdf.AddUTF8Font("Roboto", "B", "Roboto-Bold.ttf")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(false, 10)
	pdf.AddPage()
	return pdf
}

func (g *PDFGenerator) GenerateProfitReport(data []ProfitItem, from, to time.Time, taxRate decimal.Decimal) ([]byte, error) {
	pdf := g.initPDF("L")
	g.drawReportHeader(pdf, "Отчет о валовой прибыли", from, to, taxRate)

	headers := []string{"Артикул", "Товар", "Кол-во", "Выручка", "Себест.", "Прибыль", "Налог", "Чистая", "Рент.%"}
	widths := []float64{35, 75, 15, 25, 25, 25, 25, 25, 20}
	aligns := []string{"L", "L", "R", "R", "R", "R", "R", "R", "R"}
	wrapCols := []bool{false, true, false, false, false, false, false, false, false}

	g.drawTableHeader(pdf, headers, widths, aligns)

	pdf.SetFont("Roboto", "", 9)
	pdf.SetTextColor(0, 0, 0)
	tRev, tCost, tGross, tTax, tNet := decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero

	for _, item := range data {
		tRev = tRev.Add(item.SalesTotal)
		tCost = tCost.Add(item.CostTotal)
		tGross = tGross.Add(item.GrossProfit)
		tTax = tTax.Add(item.EstimatedTax)
		tNet = tNet.Add(item.NetProfit)

		rowValues := []string{
			cleanString(item.SKU), cleanString(item.ProductName), item.QuantitySold.StringFixed(0),
			fmtMoney(item.SalesTotal), fmtMoney(item.CostTotal), fmtMoney(item.GrossProfit),
			fmtMoney(item.EstimatedTax), fmtMoney(item.NetProfit), item.Profitability.StringFixed(1),
		}
		g.drawSmartRow(pdf, widths, aligns, wrapCols, rowValues)
	}
	g.drawTotalRow(pdf, widths, []string{"ИТОГО:", fmtMoney(tRev), fmtMoney(tCost), fmtMoney(tGross), fmtMoney(tTax), fmtMoney(tNet), calcAvgRent(tRev, tGross)}, []int{0, 1, 2})
	g.drawFooter(pdf)
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (g *PDFGenerator) GenerateStockReport(data []StockItem, date time.Time) ([]byte, error) {
	pdf := g.initPDF("L")
	g.drawReportHeaderSimple(pdf, "Оценка склада (Остатки)", fmt.Sprintf("На дату: %s", date.Format("02.01.2006 15:04")))

	headers := []string{"Склад", "Категория", "Артикул", "Товар", "Ед.", "Остаток", "Цена", "Сумма"}
	widths := []float64{40, 35, 30, 80, 15, 20, 25, 30} // 275
	aligns := []string{"L", "L", "L", "L", "C", "R", "R", "R"}
	wrapCols := []bool{true, true, false, true, false, false, false, false}

	g.drawTableHeader(pdf, headers, widths, aligns)

	pdf.SetFont("Roboto", "", 9)
	pdf.SetTextColor(0, 0, 0)
	totalQty := decimal.Zero
	totalSum := decimal.Zero

	for _, item := range data {
		totalQty = totalQty.Add(item.Quantity)
		totalSum = totalSum.Add(item.TotalValue)
		rowValues := []string{
			cleanString(item.WarehouseName), cleanString(item.Category), cleanString(item.SKU),
			cleanString(item.ProductName), cleanString(item.Unit), item.Quantity.StringFixed(2),
			fmtMoney(item.AvgPurchasePrice), fmtMoney(item.TotalValue),
		}
		g.drawSmartRow(pdf, widths, aligns, wrapCols, rowValues)
	}
	g.drawTotalRow(pdf, widths, []string{"ИТОГО:", totalQty.StringFixed(2), "", fmtMoney(totalSum)}, []int{0, 1, 2, 3, 4})
	g.drawFooter(pdf)
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (g *PDFGenerator) GenerateMovementsReport(data []MovementItem, from, to time.Time) ([]byte, error) {
	pdf := g.initPDF("L")
	g.drawReportHeaderSimple(pdf, "Ведомость движений товаров", fmt.Sprintf("Период: %s - %s", from.Format("02.01.2006"), to.Format("02.01.2006")))

	headers := []string{"Дата", "Документ", "Склад", "Товар (SKU)", "Ед.", "Приход", "Расход"}
	widths := []float64{28, 35, 45, 107, 15, 20, 20}
	aligns := []string{"C", "L", "L", "L", "C", "R", "R"}
	wrapCols := []bool{false, false, false, true, false, false, false}

	g.drawTableHeader(pdf, headers, widths, aligns)

	pdf.SetFont("Roboto", "", 8)
	pdf.SetTextColor(0, 0, 0)
	totalIn, totalOut := decimal.Zero, decimal.Zero

	for _, item := range data {
		totalIn = totalIn.Add(item.QuantityIn)
		totalOut = totalOut.Add(item.QuantityOut)
		inStr := ""
		if item.QuantityIn.IsPositive() {
			inStr = item.QuantityIn.StringFixed(2)
		}
		outStr := ""
		if item.QuantityOut.IsPositive() {
			outStr = item.QuantityOut.StringFixed(2)
		}
		docStr := fmt.Sprintf("%s %s", getDocTypeShort(item.DocumentType), item.DocumentNumber)
		prodStr := fmt.Sprintf("%s (%s)", cleanString(item.ProductName), cleanString(item.SKU))

		rowValues := []string{
			item.Date.Format("02.01.06 15:04"), cleanString(docStr), cleanString(item.WarehouseName),
			prodStr, cleanString(item.Unit), inStr, outStr,
		}
		g.drawSmartRow(pdf, widths, aligns, wrapCols, rowValues)
	}
	g.drawTotalRow(pdf, widths, []string{"ИТОГО:", totalIn.StringFixed(2), totalOut.StringFixed(2)}, []int{0, 1, 2, 3, 4})
	g.drawFooter(pdf)
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (g *PDFGenerator) GenerateCustomerReport(data []CustomerReportItem, from, to time.Time) ([]byte, error) {
	pdf := g.initPDF("P")
	g.drawReportHeaderSimple(pdf, "Продажи по клиентам", fmt.Sprintf("Период: %s - %s", from.Format("02.01.2006"), to.Format("02.01.2006")))

	headers := []string{"Контрагент", "Кол-во сделок", "Сумма покупок"}
	widths := []float64{100, 40, 50}
	aligns := []string{"L", "C", "R"}
	wrapCols := []bool{true, false, false}

	g.drawTableHeader(pdf, headers, widths, aligns)

	pdf.SetFont("Roboto", "", 10)
	pdf.SetTextColor(0, 0, 0)
	totalOps := int64(0)
	totalRev := decimal.Zero

	for _, item := range data {
		totalOps += item.OperationsCount
		totalRev = totalRev.Add(item.TotalRevenue)
		rowValues := []string{
			cleanString(item.CounterpartyName), fmt.Sprintf("%d", item.OperationsCount), fmtMoney(item.TotalRevenue),
		}
		g.drawSmartRow(pdf, widths, aligns, wrapCols, rowValues)
	}
	g.drawTotalRow(pdf, widths, []string{"ИТОГО:", fmt.Sprintf("%d", totalOps), fmtMoney(totalRev)}, []int{0})
	g.drawFooter(pdf)
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (g *PDFGenerator) GenerateABCReport(data []ABCItem, from, to time.Time) ([]byte, error) {
	pdf := g.initPDF("P")

	period := fmt.Sprintf("Период: %s - %s", from.Format("02.01.2006"), to.Format("02.01.2006"))
	g.drawReportHeaderSimple(pdf, "Рейтинг продаж (ABC-анализ)", period)

	headers := []string{"Артикул", "Товар", "Продано", "Выручка", "Доля %", "Класс"}
	widths := []float64{35, 80, 20, 30, 15, 10}
	aligns := []string{"L", "L", "R", "R", "R", "C"}
	wrapCols := []bool{false, true, false, false, false, false}

	g.drawTableHeader(pdf, headers, widths, aligns)

	pdf.SetFont("Roboto", "", 9)
	pdf.SetTextColor(0, 0, 0)

	totalRev := decimal.Zero

	for _, item := range data {
		totalRev = totalRev.Add(item.Revenue)

		rowValues := []string{
			cleanString(item.SKU),
			cleanString(item.ProductName),
			item.QuantitySold.StringFixed(0),
			fmtMoney(item.Revenue),
			item.SharePercent.StringFixed(2),
			item.Class,
		}
		g.drawSmartRow(pdf, widths, aligns, wrapCols, rowValues)
	}

	g.drawTotalRow(pdf, widths, []string{
		"ИТОГО:",
		fmtMoney(totalRev),
		"", "",
	}, []int{0, 1, 2})

	g.drawFooter(pdf)
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (g *PDFGenerator) drawSmartRow(pdf *fpdf.Fpdf, widths []float64, aligns []string, wrapCols []bool, data []string) {
	const lineHeight = 3.5
	const cellPadding = 1.0
	maxLines := 1
	for i, text := range data {
		if !wrapCols[i] {
			continue
		}
		lines := pdf.SplitLines([]byte(text), widths[i]-1.0)
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	rowHeight := (float64(maxLines) * lineHeight) + (cellPadding * 2)

	_, pageH := pdf.GetPageSize()
	_, _, _, bottomMargin := pdf.GetMargins()
	if pdf.GetY()+rowHeight > pageH-bottomMargin {
		pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+sumWidths(widths), pdf.GetY())
		pdf.AddPage()
		pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+sumWidths(widths), pdf.GetY())
	}
	startY := pdf.GetY()
	currentX := pdf.GetX()

	for i, text := range data {
		pdf.Line(currentX, startY, currentX, startY+rowHeight)
		thisLines := 1
		if wrapCols[i] {
			lines := pdf.SplitLines([]byte(text), widths[i]-1.0)
			thisLines = len(lines)
			if thisLines == 0 {
				thisLines = 1
			}
		}
		textHeight := float64(thisLines) * lineHeight
		verticalOffset := (rowHeight - textHeight) / 2.0
		pdf.SetXY(currentX, startY+verticalOffset)

		if wrapCols[i] {
			pdf.MultiCell(widths[i], lineHeight, text, "", aligns[i], false)
		} else {
			pdf.CellFormat(widths[i], lineHeight, text, "", 0, aligns[i], false, 0, "")
		}
		currentX += widths[i]
	}
	pdf.Line(currentX, startY, currentX, startY+rowHeight)
	pdf.Line(10, startY+rowHeight, currentX, startY+rowHeight)
	pdf.SetXY(10, startY+rowHeight)
}

func (g *PDFGenerator) drawTableHeader(pdf *fpdf.Fpdf, headers []string, widths []float64, aligns []string) {
	pdf.SetFont("Roboto", "B", 10)
	pdf.SetFillColor(ColorGrey[0], ColorGrey[1], ColorGrey[2])
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
	startX := pdf.GetX()
	startY := pdf.GetY()
	rowHeight := 8.0
	currentX := startX
	pdf.Line(currentX, startY, currentX+sumWidths(widths), startY)
	for i, h := range headers {
		pdf.Line(currentX, startY, currentX, startY+rowHeight)
		pdf.SetXY(currentX, startY)
		pdf.CellFormat(widths[i], rowHeight, h, "", 0, aligns[i], true, 0, "")
		currentX += widths[i]
	}
	pdf.Line(currentX, startY, currentX, startY+rowHeight)
	pdf.Line(startX, startY+rowHeight, currentX, startY+rowHeight)
	pdf.Ln(rowHeight)
}

func (g *PDFGenerator) drawTotalRow(pdf *fpdf.Fpdf, widths []float64, values []string, mergeCols []int) {
	pdf.SetFont("Roboto", "B", 9)
	pdf.SetFillColor(ColorGrey[0], ColorGrey[1], ColorGrey[2])
	rowHeight := 7.0
	startX := pdf.GetX()
	startY := pdf.GetY()
	currentX := startX
	totalWidth := 0.0
	for _, idx := range mergeCols {
		totalWidth += widths[idx]
	}

	pdf.Line(currentX, startY, currentX, startY+rowHeight)
	pdf.SetXY(currentX, startY)
	pdf.CellFormat(totalWidth, rowHeight, values[0], "", 0, "R", true, 0, "")
	currentX += totalWidth
	valIndex := 1
	for i := 0; i < len(widths); i++ {
		isMerged := false
		for _, m := range mergeCols {
			if i == m {
				isMerged = true
				break
			}
		}
		if isMerged {
			continue
		}
		pdf.Line(currentX, startY, currentX, startY+rowHeight)
		val := ""
		if valIndex < len(values) {
			val = values[valIndex]
			valIndex++
		}
		pdf.SetXY(currentX, startY)
		pdf.CellFormat(widths[i], rowHeight, val, "", 0, "R", true, 0, "")
		currentX += widths[i]
	}
	pdf.Line(currentX, startY, currentX, startY+rowHeight)
	pdf.Line(startX, startY+rowHeight, currentX, startY+rowHeight)
	pdf.Ln(10)
}

func (g *PDFGenerator) drawReportHeader(pdf *fpdf.Fpdf, title string, from, to time.Time, taxRate decimal.Decimal) {
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Roboto", "B", 16)
	pdf.Cell(0, 8, title)
	pdf.Ln(6)
	pdf.SetFont("Roboto", "", 10)
	pdf.Write(5, fmt.Sprintf("Период: %s - %s", from.Format("02.01.2006"), to.Format("02.01.2006")))
	pdf.Ln(5)
	pdf.Write(5, "Организация: Мой Склад (ООО)")
	pdf.Ln(5)
	if taxRate.IsPositive() {
		pdf.SetFont("Roboto", "", 9)
		pdf.Write(5, fmt.Sprintf("Налогообложение: Ставка %s%% (Доходы минус Расходы)", taxRate.String()))
		pdf.Ln(5)
	}
	pdf.Ln(5)
}

func (g *PDFGenerator) drawReportHeaderSimple(pdf *fpdf.Fpdf, title, subtitle string) {
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Roboto", "B", 16)
	pdf.Cell(0, 8, title)
	pdf.Ln(8)
	pdf.SetFont("Roboto", "", 10)
	pdf.Write(5, subtitle)
	pdf.Ln(5)
	pdf.Write(5, "Организация: Мой Склад (ООО)")
	pdf.Ln(8)
}

func (g *PDFGenerator) drawFooter(pdf *fpdf.Fpdf) {
	pdf.SetY(-20)
	pdf.SetFont("Roboto", "", 8)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(20, 5, "Сформировал:")
	pdf.Cell(60, 5, "_________________________")
	pdf.Cell(20, 5, "Подпись:")
	pdf.Cell(60, 5, "_________________________")
	pdf.Ln(6)
	pdf.SetFont("Roboto", "", 7)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, fmt.Sprintf("Документ создан в FlowKeeper: %s | Страница %d", time.Now().Format("02.01.2006 15:04"), pdf.PageNo()))
}

func cleanString(s string) string {
	s = strings.ReplaceAll(s, "\u00A0", " ")
	s = strings.Map(func(r rune) rune {
		if r == 0 || r == 0xFFFD {
			return -1
		}
		return r
	}, s)
	return s
}
func fmtMoney(d decimal.Decimal) string { v, _ := d.Float64(); return fmt.Sprintf("%.2f", v) }
func getDocTypeShort(t string) string {
	switch t {
	case "INCOME":
		return "Прих"
	case "OUTCOME":
		return "Расх"
	case "ORDER":
		return "Заказ"
	case "TRANSFER":
		return "Перем"
	default:
		return t
	}
}
func calcAvgRent(rev, gross decimal.Decimal) string {
	if !rev.IsZero() {
		return gross.Div(rev).Mul(decimal.NewFromInt(100)).StringFixed(1) + "%"
	}
	return "0.0%"
}
func sumWidths(w []float64) float64 {
	sum := 0.0
	for _, v := range w {
		sum += v
	}
	return sum
}
