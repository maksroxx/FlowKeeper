package analytics

import (
	"github.com/shopspring/decimal"
)

type Service interface {
	GetDashboardData(warehouseID *uint) (*DashboardData, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetDashboardData(warehouseID *uint) (*DashboardData, error) {
	// 1. Общий остаток (сумма количеств)
	totalStock, _, err := s.repo.GetTotalStock(warehouseID)
	if err != nil {
		return nil, err
	}

	// 2. Здоровье склада (Всего позиций / В наличии / Дефицит)
	totalVariants, inStock, lowStock, err := s.repo.GetInventoryHealth(warehouseID)
	if err != nil {
		return nil, err
	}

	// 3. Активность (операций за 30 дней, приход/расход за сегодня)
	recentOps, inToday, outToday, err := s.repo.GetActivityStats(warehouseID, 30)
	if err != nil {
		return nil, err
	}

	// 4. Подготовка данных для графика
	movementsRaw, err := s.repo.GetChartData(warehouseID, 30)
	if err != nil {
		return nil, err
	}

	// Агрегация графика в памяти (группировка по дате и типу)
	chartMap := make(map[string]map[string]decimal.Decimal)

	for _, m := range movementsRaw {
		date := m.CreatedAt.Format("2006-01-02")
		if _, ok := chartMap[date]; !ok {
			chartMap[date] = map[string]decimal.Decimal{
				"INCOME":  decimal.Zero,
				"OUTCOME": decimal.Zero,
			}
		}

		qty := m.Quantity.Abs()
		mType := "INCOME"
		if m.Type == "OUTCOME" {
			mType = "OUTCOME"
		}

		if m.Type == "INCOME" || m.Type == "OUTCOME" {
			chartMap[date][mType] = chartMap[date][mType].Add(qty)
		}
	}

	var chartData []ChartPoint
	for date, types := range chartMap {
		if !types["INCOME"].IsZero() {
			chartData = append(chartData, ChartPoint{Date: date, Value: types["INCOME"], Type: "INCOME"})
		}
		if !types["OUTCOME"].IsZero() {
			chartData = append(chartData, ChartPoint{Date: date, Value: types["OUTCOME"], Type: "OUTCOME"})
		}
	}

	// 5. Последние движения (Получаем готовые DTO с именами через JOIN)
	recentMovements, err := s.repo.GetRecentMovements(warehouseID, 5)
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		TotalStock:       totalStock,
		TotalItemsCount:  totalVariants, // Используем общее кол-во вариантов как "Всего номенклатуры"
		TotalVariants:    totalVariants,
		ItemsInStock:     inStock,
		LowStockCount:    lowStock,
		RecentOperations: recentOps,
		IncomingToday:    inToday,
		OutgoingToday:    outToday,
		ChartData:        chartData,
		RecentMovements:  recentMovements,
	}, nil
}
