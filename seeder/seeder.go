package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	cfgg "github.com/maksroxx/flowkeeper/internal/config"
	"github.com/maksroxx/flowkeeper/internal/db"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/config"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/service"
)

type seeder struct {
	db          *gorm.DB
	productSvc  service.ProductService
	variantSvc  service.VariantService
	categorySvc service.CategoryService
	unitSvc     service.UnitService
	whSvc       service.WarehouseService
	docSvc      service.DocumentService
	cpSvc       service.CounterpartyService
	ptSvc       service.PriceTypeService
	charSvc     service.CharacteristicService
}

func main() {
	log.Println("Запуск сидера для наполнения базы данных...")

	cfg, err := cfgg.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}
	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	s := newSeeder(database)

	log.Println("Очистка старых данных...")
	s.clearData()

	log.Println("Создание справочников...")
	categories, units, warehouses := s.createDirectories()

	log.Println("Создание номенклатуры и вариантов...")
	variants := s.createProductsAndVariants(categories, units)

	log.Println("Создание и проведение документа 'Приход'...")
	s.createIncomeDocument(warehouses[0], variants)

	log.Println("Наполнение базы данных успешно завершено!")
}

func newSeeder(db *gorm.DB) *seeder {
	// --- repositories ---
	productRepo := repository.NewProductRepository(db)
	variantRepo := repository.NewVariantRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	unitRepo := repository.NewUnitRepository(db)
	whRepo := repository.NewWarehouseRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	historyRepo := repository.NewDocumentHistoryRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	movementRepo := repository.NewStockMovementRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	lotRepo := repository.NewLotRepository()
	priceRepo := repository.NewPriceRepository(db)
	seqRepo := repository.NewSequenceRepository()
	counterpartyRepo := repository.NewCounterpartyRepository(db)
	priceTypeRepo := repository.NewPriceTypeRepository(db)
	charactRepo := repository.NewCharacteristicRepository(db)
	txManager := repository.NewTxManager(db)

	// --- services ---
	stockCfg, err := config.LoadStockConfig("./config/stock_config.yml")
	if err != nil {
		panic(fmt.Sprintf("failed to load stock module config: %v", err))
	}

	strategyFactory := service.NewStrategyFactory(balanceRepo, movementRepo, lotRepo)

	inventorySvc := service.NewInventoryService(
		strategyFactory,
		reservationRepo,
		balanceRepo,
		stockCfg,
		variantRepo,
		productRepo,
		unitRepo,
	)

	priceSvc := service.NewPriceService(priceRepo)
	seqSvc := service.NewSequenceService(seqRepo, txManager)

	docSvc := service.NewDocumentService(
		docRepo, historyRepo, inventorySvc, priceSvc, seqSvc, txManager,
		variantRepo, productRepo, whRepo, counterpartyRepo, priceTypeRepo,
	)

	return &seeder{
		db:          db,
		productSvc:  service.NewProductService(productRepo),
		variantSvc:  service.NewVariantService(variantRepo, productRepo, unitRepo),
		categorySvc: service.NewCategoryService(categoryRepo),
		unitSvc:     service.NewUnitService(unitRepo),
		whSvc:       service.NewWarehouseService(whRepo),
		docSvc:      docSvc,
		cpSvc:       service.NewCounterpartyService(counterpartyRepo),
		ptSvc:       service.NewPriceTypeService(priceTypeRepo),
		charSvc:     service.NewCharacteristicService(charactRepo),
	}
}

func (s *seeder) clearData() {
	s.db.Exec("DELETE FROM item_prices")
	s.db.Exec("DELETE FROM stock_reservations")
	s.db.Exec("DELETE FROM stock_lots")
	s.db.Exec("DELETE FROM stock_movements")
	s.db.Exec("DELETE FROM stock_balances")
	s.db.Exec("DELETE FROM document_items")
	s.db.Exec("DELETE FROM document_histories")
	s.db.Exec("DELETE FROM documents")
	s.db.Exec("DELETE FROM document_sequences")
	s.db.Exec("DELETE FROM variants")
	s.db.Exec("DELETE FROM products")
	s.db.Exec("DELETE FROM characteristic_values")
	s.db.Exec("DELETE FROM characteristic_types")
	s.db.Exec("DELETE FROM categories")
	s.db.Exec("DELETE FROM units")
	s.db.Exec("DELETE FROM warehouses")
	s.db.Exec("DELETE FROM counterparties")
	s.db.Exec("DELETE FROM price_types")
}

func (s *seeder) createDirectories() (map[string]models.Category, map[string]models.Unit, []models.Warehouse) {
	categories := make(map[string]models.Category)
	catNames := []string{"Ноутбуки", "Смартфоны", "Аксессуары", "Одежда"}
	for _, name := range catNames {
		cat, _ := s.categorySvc.Create(name)
		categories[name] = *cat
	}

	units := make(map[string]models.Unit)
	unitNames := []string{"шт", "кг", "пара"}
	for _, name := range unitNames {
		unit, _ := s.unitSvc.Create(name)
		units[name] = *unit
	}

	s.cpSvc.Create(&models.Counterparty{Name: "Розничный покупатель"})
	s.cpSvc.Create(&models.Counterparty{Name: "ООО 'Поставщик №1'"})

	s.ptSvc.Create("Закупочная")
	s.ptSvc.Create("Розничная")

	warehouses := []models.Warehouse{}
	wh1, _ := s.whSvc.Create("Основной склад (Москва)", "ул. Тверская, д.1")
	wh2, _ := s.whSvc.Create("Склад (СПб)", "Невский пр-т, д. 10")
	warehouses = append(warehouses, *wh1, *wh2)

	return categories, units, warehouses
}

func (s *seeder) createProductsAndVariants(categories map[string]models.Category, units map[string]models.Unit) []models.Variant {
	var variants []models.Variant

	// Ноутбуки
	p1, _ := s.productSvc.Create(&models.Product{Name: "Ноутбук Pro 16", CategoryID: categories["Ноутбуки"].ID})
	v1, _ := s.variantSvc.Create(&models.Variant{ProductID: p1.ID, SKU: "NBP-16-512", UnitID: units["шт"].ID, Characteristics: models.CharacteristicsMap{"Память": "512Гб", "Цвет": "Серый"}})
	v2, _ := s.variantSvc.Create(&models.Variant{ProductID: p1.ID, SKU: "NBP-16-1TB", UnitID: units["шт"].ID, Characteristics: models.CharacteristicsMap{"Память": "1Тб", "Цвет": "Серый"}})

	// Смартфоны
	p2, _ := s.productSvc.Create(&models.Product{Name: "Смартфон X1", CategoryID: categories["Смартфоны"].ID})
	v3, _ := s.variantSvc.Create(&models.Variant{ProductID: p2.ID, SKU: "PH-X1-BLK", UnitID: units["шт"].ID, Characteristics: models.CharacteristicsMap{"Цвет": "Черный"}})
	v4, _ := s.variantSvc.Create(&models.Variant{ProductID: p2.ID, SKU: "PH-X1-WHT", UnitID: units["шт"].ID, Characteristics: models.CharacteristicsMap{"Цвет": "Белый"}})

	// Аксессуары
	p3, _ := s.productSvc.Create(&models.Product{Name: "Беспроводная мышь", CategoryID: categories["Аксессуары"].ID})
	v5, _ := s.variantSvc.Create(&models.Variant{ProductID: p3.ID, SKU: "MOUSE-WL-01", UnitID: units["шт"].ID})

	p4, _ := s.productSvc.Create(&models.Product{Name: "Клавиатура механическая", CategoryID: categories["Аксессуары"].ID})
	v6, _ := s.variantSvc.Create(&models.Variant{ProductID: p4.ID, SKU: "KEY-MECH-RGB", UnitID: units["шт"].ID})

	// Одежда
	p5, _ := s.productSvc.Create(&models.Product{Name: "Футболка Поло", CategoryID: categories["Одежда"].ID})
	v7, _ := s.variantSvc.Create(&models.Variant{ProductID: p5.ID, SKU: "POLO-BL-S", UnitID: units["шт"].ID, Characteristics: models.CharacteristicsMap{"Цвет": "Синий", "Размер": "S"}})
	v8, _ := s.variantSvc.Create(&models.Variant{ProductID: p5.ID, SKU: "POLO-BL-M", UnitID: units["шт"].ID, Characteristics: models.CharacteristicsMap{"Цвет": "Синий", "Размер": "M"}})

	p6, _ := s.productSvc.Create(&models.Product{Name: "Кроссовки", CategoryID: categories["Одежда"].ID})
	v9, _ := s.variantSvc.Create(&models.Variant{ProductID: p6.ID, SKU: "SHOE-RUN-42", UnitID: units["пара"].ID, Characteristics: models.CharacteristicsMap{"Размер": "42"}})
	v10, _ := s.variantSvc.Create(&models.Variant{ProductID: p6.ID, SKU: "SHOE-RUN-43", UnitID: units["пара"].ID, Characteristics: models.CharacteristicsMap{"Размер": "43"}})

	variants = append(variants, *v1, *v2, *v3, *v4, *v5, *v6, *v7, *v8, *v9, *v10)
	return variants
}

func (s *seeder) createIncomeDocument(warehouse models.Warehouse, variants []models.Variant) {
	docItems := make([]models.DocumentItem, len(variants))

	for i, v := range variants {
		qty, _ := decimal.NewFromString(fmt.Sprintf("%d.%d", 10+i*2, i))
		price, _ := decimal.NewFromString(fmt.Sprintf("%d", 1000+i*150))
		docItems[i] = models.DocumentItem{
			VariantID: v.ID,
			Quantity:  qty,
			Price:     &price,
		}
	}

	whID := warehouse.ID
	doc, err := s.docSvc.Create(&models.Document{
		Type:        "INCOME",
		WarehouseID: &whID,
		CreatedAt:   time.Now(),
		Comment:     "Начальное наполнение базы данных",
		Items:       docItems,
	})
	if err != nil {
		log.Fatalf("Не удалось создать документ прихода: %v", err)
	}

	err = s.docSvc.Post(doc.ID)
	if err != nil {
		log.Fatalf("Не удалось провести документ прихода: %v", err)
	}
}
