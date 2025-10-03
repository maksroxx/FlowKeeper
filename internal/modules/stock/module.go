package stock

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/config"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/handler"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type Module struct{}

func NewModule() *Module { return &Module{} }

func (m *Module) Name() string { return "stock" }

func (m *Module) RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	grp := r.Group("/api/v1/stock")

	stockCfg, err := config.LoadStockConfig("./config/stock_config.yml")
	if err != nil {
		panic(fmt.Sprintf("failed to load stock module config: %v", err))
	}

	// --- repositories ---
	productRepo := repository.NewProductRepository(db)
	variantRepo := repository.NewVariantRepository(db)
	charactRepo := repository.NewCharacteristicRepository(db)
	priceTypeRepo := repository.NewPriceTypeRepository(db)
	catRepo := repository.NewCategoryRepository(db)
	cpRepo := repository.NewCounterpartyRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	movRepo := repository.NewStockMovementRepository(db)
	unitRepo := repository.NewUnitRepository(db)
	whRepo := repository.NewWarehouseRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	historyRepo := repository.NewDocumentHistoryRepository(db)
	priceRepo := repository.NewPriceRepository(db)
	seqRepo := repository.NewSequenceRepository()
	reservRepo := repository.NewReservationRepository(db)
	txManager := repository.NewTxManager(db)
	lotRepo := repository.NewLotRepository()

	// --- services ---
	productSvc := service.NewProductService(productRepo)
	variantSvc := service.NewVariantService(variantRepo, productRepo, unitRepo)
	charactSvc := service.NewCharacteristicService(charactRepo)
	priceTypeSvc := service.NewPriceTypeService(priceTypeRepo)
	priceSvc := service.NewPriceService(priceRepo)
	seqSvc := service.NewSequenceService(seqRepo, txManager)
	catSvc := service.NewCategoryService(catRepo)
	cpSvc := service.NewCounterpartyService(cpRepo)
	strategyFactory := service.NewStrategyFactory(balanceRepo, movRepo, lotRepo)
	inventorySvc := service.NewInventoryService(strategyFactory, reservRepo, balanceRepo, stockCfg, variantRepo, productRepo, unitRepo)
	docSvc := service.NewDocumentService(
		docRepo, historyRepo,
		inventorySvc, priceSvc,
		seqSvc, txManager,
		variantRepo, productRepo,
		whRepo, cpRepo,
		priceTypeRepo,
	)
	movSvc := service.NewStockMovementService(movRepo, docRepo, variantRepo, productRepo, whRepo)
	unitSvc := service.NewUnitService(unitRepo)
	whSvc := service.NewWarehouseService(whRepo)

	// --- handlers ---
	handler.NewProductHandler(productSvc).Register(grp)
	handler.NewCharacteristicHandler(charactSvc).Register(grp)
	handler.NewVariantHandler(variantSvc).Register(grp)
	handler.NewPriceTypeHandler(priceTypeSvc).Register(grp)
	handler.NewPriceHandler(priceSvc).Register(grp)
	handler.NewCategoryHandler(catSvc).Register(grp)
	handler.NewCounterpartyHandler(cpSvc).Register(grp)
	handler.NewDocumentHandler(docSvc).Register(grp)
	handler.NewMovementHandler(movSvc).Register(grp)
	handler.NewUnitHandler(unitSvc).Register(grp)
	handler.NewWarehouseHandler(whSvc).Register(grp)
	handler.NewBalanceHandler(inventorySvc).Register(grp)
}

func (m *Module) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Product{},
		&models.Variant{},
		&models.CharacteristicType{},
		&models.CharacteristicValue{},
		&models.StockReservation{},
		&models.StockLot{},

		&models.Category{},
		&models.Counterparty{},
		&models.Document{},
		&models.DocumentItem{},
		// &models.Item{},
		&models.ItemPrice{},
		&models.PriceType{},
		&models.StockMovement{},
		&models.Unit{},
		&models.Warehouse{},
		&models.StockBalance{},
		&models.DocumentHistory{},
		&models.DocumentSequence{},
	)
}
