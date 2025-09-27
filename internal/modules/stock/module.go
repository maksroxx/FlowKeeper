package stock

import (
	"github.com/gin-gonic/gin"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/handler"
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
	"gorm.io/gorm"
)

type Module struct{}

func NewModule() *Module { return &Module{} }

func (m *Module) Name() string { return "stock" }

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	grp := r.Group("/api/v1/stock")

	// repositories
	catRepo := repository.NewCategoryRepository(db)
	cpRepo := repository.NewCounterpartyRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	itemRepo := repository.NewItemRepository(db)
	movRepo := repository.NewStockMovementRepository(db)
	unitRepo := repository.NewUnitRepository(db)
	whRepo := repository.NewWarehouseRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	historyRepo := repository.NewDocumentHistoryRepository(db)
	txManager := repository.NewTxManager(db)

	// services
	catSvc := service.NewCategoryService(catRepo)
	cpSvc := service.NewCounterpartyService(cpRepo)
	inventorySvc := service.NewInventoryService(balanceRepo, movRepo, txManager)
	docSvc := service.NewDocumentService(docRepo, historyRepo, inventorySvc, txManager)
	itemSvc := service.NewItemService(itemRepo)
	movSvc := service.NewStockMovementService(movRepo)
	unitSvc := service.NewUnitService(unitRepo)
	whSvc := service.NewWarehouseService(whRepo)

	// handlers
	handler.NewCategoryHandler(catSvc).Register(grp)
	handler.NewCounterpartyHandler(cpSvc).Register(grp)
	handler.NewDocumentHandler(docSvc).Register(grp)
	handler.NewItemHandler(itemSvc).Register(grp)
	handler.NewMovementHandler(movSvc).Register(grp)
	handler.NewUnitHandler(unitSvc).Register(grp)
	handler.NewWarehouseHandler(whSvc).Register(grp)
	handler.NewBalanceHandler(inventorySvc).Register(grp)
}

func (m *Module) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&stock.Category{},
		&stock.Counterparty{},
		&stock.Document{},
		&stock.DocumentItem{},
		&stock.Item{},
		&stock.StockMovement{},
		&stock.Unit{},
		&stock.Warehouse{},
		&stock.StockBalance{},
		&stock.DocumentHistory{},
	)
}
