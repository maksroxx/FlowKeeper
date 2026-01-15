package reports

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Service struct {
	repo     Repository
	pdfGen   *PDFGenerator
	excelGen *ExcelGenerator
	csvGen   *CSVGenerator
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:     repo,
		pdfGen:   NewPDFGenerator("assets/fonts/"),
		excelGen: NewExcelGenerator(),
		csvGen:   NewCSVGenerator(),
	}
}

func (s *Service) GenerateReport(req ReportRequest) ([]byte, string, error) {
	switch req.Type {
	case "profit":
		return s.generateProfit(req)
	case "stock":
		return s.generateStock(req)
	case "movements":
		return s.generateMovements(req)
	case "customers":
		return s.generateCustomers(req)
	default:
		return nil, "", fmt.Errorf("unknown report type: %s", req.Type)
	}
}

func (s *Service) generateProfit(req ReportRequest) ([]byte, string, error) {
	records, err := s.repo.GetFIFOProfitData(req.DateFrom, req.DateTo, req.WarehouseID)
	if err != nil {
		return nil, "", err
	}

	var rows []ProfitItem
	if req.TaxRate.IsNegative() {
		req.TaxRate = decimal.Zero
	}
	taxMultiplier := req.TaxRate.Div(decimal.NewFromInt(100))

	for _, rec := range records {
		gross := rec.Revenue.Sub(rec.Cost)
		tax := decimal.Zero
		if gross.IsPositive() {
			tax = gross.Mul(taxMultiplier)
		}
		net := gross.Sub(tax)
		rent := decimal.Zero
		if !rec.Revenue.IsZero() {
			rent = gross.Div(rec.Revenue).Mul(decimal.NewFromInt(100))
		}

		rows = append(rows, ProfitItem{
			SKU: rec.SKU, ProductName: rec.ProductName, QuantitySold: rec.Quantity,
			SalesTotal: rec.Revenue, CostTotal: rec.Cost, GrossProfit: gross,
			EstimatedTax: tax, NetProfit: net, Profitability: rent,
		})
	}

	switch req.Format {
	case "excel", "xlsx":
		b, err := s.excelGen.GenerateProfit(rows)
		return b, "xlsx", err
	case "csv":
		b, err := s.csvGen.GenerateProfit(rows)
		return b, "csv", err
	default:
		b, err := s.pdfGen.GenerateProfitReport(rows, req.DateFrom, req.DateTo, req.TaxRate)
		return b, "pdf", err
	}
}

func (s *Service) generateStock(req ReportRequest) ([]byte, string, error) {
	data, err := s.repo.GetStockData(req.WarehouseID)
	if err != nil {
		return nil, "", err
	}

	for i := range data {
		if data[i].Quantity.IsPositive() {
			data[i].AvgPurchasePrice = data[i].TotalValue.Div(data[i].Quantity)
		}
	}

	switch req.Format {
	case "excel", "xlsx":
		b, err := s.excelGen.GenerateStock(data)
		return b, "xlsx", err
	case "csv":
		b, err := s.csvGen.GenerateStock(data)
		return b, "csv", err
	default:
		b, err := s.pdfGen.GenerateStockReport(data, req.DateTo)
		return b, "pdf", err
	}
}

func (s *Service) generateMovements(req ReportRequest) ([]byte, string, error) {
	data, err := s.repo.GetMovementsData(req.DateFrom, req.DateTo, req.WarehouseID)
	if err != nil {
		return nil, "", err
	}

	switch req.Format {
	case "excel", "xlsx":
		b, err := s.excelGen.GenerateMovements(data)
		return b, "xlsx", err
	case "csv":
		b, err := s.csvGen.GenerateMovements(data)
		return b, "csv", err
	default:
		b, err := s.pdfGen.GenerateMovementsReport(data, req.DateFrom, req.DateTo)
		return b, "pdf", err
	}
}

func (s *Service) generateCustomers(req ReportRequest) ([]byte, string, error) {
	data, err := s.repo.GetCustomerData(req.DateFrom, req.DateTo)
	if err != nil {
		return nil, "", err
	}

	switch req.Format {
	case "excel", "xlsx":
		b, err := s.excelGen.GenerateCustomers(data)
		return b, "xlsx", err

	case "csv":
		b, err := s.csvGen.GenerateCustomers(data)
		return b, "csv", err

	default:
		b, err := s.pdfGen.GenerateCustomerReport(data, req.DateFrom, req.DateTo)
		return b, "pdf", err
	}
}
