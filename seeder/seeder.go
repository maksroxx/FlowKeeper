package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
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
	s.createIncomeDocument(warehouses, variants, units)

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
		categoryRepo,
		whRepo,
	)

	priceSvc := service.NewPriceService(priceRepo)
	seqSvc := service.NewSequenceService(seqRepo, txManager)
	unitSvc := service.NewUnitService(unitRepo)
	productSvc := service.NewProductService(productRepo, variantRepo)
	variantSvc := service.NewVariantService(variantRepo, productRepo, unitRepo)

	docSvc := service.NewDocumentService(
		docRepo, historyRepo, inventorySvc, priceSvc, seqSvc, txManager,
		variantRepo, productRepo, whRepo, counterpartyRepo, priceTypeRepo,
	)

	return &seeder{
		db:          db,
		productSvc:  productSvc,
		variantSvc:  variantSvc,
		categorySvc: service.NewCategoryService(categoryRepo),
		unitSvc:     unitSvc,
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
	catNames := []string{
		"Ноутбуки", "Компьютеры и моноблоки", "Планшеты",
		"Смартфоны", "Телевизоры", "Аудиотехника", "Фото и видео",
		"Техника для кухни", "Техника для дома",
		"Мужская одежда", "Женская одежда", "Мужская обувь", "Женская обувь",
		"Спортивные товары", "Товары для туризма", "Рыбалка и охота",
		"Книги", "Канцтовары", "Товары для хобби",
		"Товары для животных", "Мебель", "Инструменты", "Автотовары",
		"Детские товары", "Косметика и парфюмерия",
	}
	for _, name := range catNames {
		cat, _ := s.categorySvc.Create(name)
		categories[name] = *cat
	}

	units := make(map[string]models.Unit)
	unitNames := []string{"шт", "кг", "пара", "л", "упак"}
	for _, name := range unitNames {
		unit, _ := s.unitSvc.Create(name)
		units[name] = *unit
	}

	s.cpSvc.Create(&models.Counterparty{Name: "Розничный покупатель"})
	s.cpSvc.Create(&models.Counterparty{Name: "ООО 'Поставщик №1'"})
	s.ptSvc.Create("Закупочная")
	s.ptSvc.Create("Розничная")

	warehouses := []models.Warehouse{}
	wh1, _ := s.whSvc.Create("Центральный склад (Москва)", "ул. Тверская, д.1")
	wh2, _ := s.whSvc.Create("Склад (Санкт-Петербург)", "Невский пр-т, д. 10")
	warehouses = append(warehouses, *wh1, *wh2)

	return categories, units, warehouses
}

func (s *seeder) createProductsAndVariants(categories map[string]models.Category, units map[string]models.Unit) []models.Variant {
	var allVariants []models.Variant

	// --- ЭЛЕКТРОНИКА ---
	s.generateVariants(&allVariants, "Ноутбук", []string{"Asus Zenbook", "Apple MacBook Pro", "HP Spectre", "Dell XPS"}, categories["Ноутбуки"].ID, units["шт"].ID, "Конфигурация", []string{"16/512GB", "16/1TB", "32/1TB"})
	s.generateVariants(&allVariants, "Моноблок", []string{"iMac 24", "Lenovo IdeaCentre"}, categories["Компьютеры и моноблоки"].ID, units["шт"].ID, "Цвет", []string{"Silver", "Space Gray", "White"})
	s.generateVariants(&allVariants, "Планшет", []string{"iPad Air", "Samsung Galaxy Tab S9", "Xiaomi Pad 6"}, categories["Планшеты"].ID, units["шт"].ID, "Память", []string{"128GB", "256GB"})
	s.generateVariants(&allVariants, "Смартфон", []string{"iPhone 17", "Samsung Galaxy S25", "Xiaomi 15"}, categories["Смартфоны"].ID, units["шт"].ID, "Цвет", []string{"Black", "White", "Blue"})
	s.generateSimpleProducts(&allVariants, []string{"Телевизор 55\" OLED", "Саундбар Dolby Atmos", "Наушники-вкладыши Pro"}, categories["Телевизоры"].ID, units["шт"].ID)

	// --- БЫТОВАЯ ТЕХНИКА ---
	s.generateSimpleProducts(&allVariants, []string{"Чайник электрический", "Кофемашина рожковая", "Микроволновая печь", "Блендер"}, categories["Техника для кухни"].ID, units["шт"].ID)
	s.generateSimpleProducts(&allVariants, []string{"Робот-пылесос", "Вертикальный пылесос", "Утюг паровой"}, categories["Техника для дома"].ID, units["шт"].ID)

	// --- ОДЕЖДА И ОБУВЬ ---
	s.generateVariants(&allVariants, "Джинсы", []string{"Levi's 501", "Wrangler Classic"}, categories["Мужская одежда"].ID, units["шт"].ID, "Размер", []string{"30/32", "32/32", "34/34"})
	s.generateVariants(&allVariants, "Футболка", []string{"Nike Sport", "Adidas Originals"}, categories["Мужская одежда"].ID, units["шт"].ID, "Размер", []string{"S", "M", "L", "XL"})
	s.generateVariants(&allVariants, "Платье", []string{"Zara Summer", "Mango Night"}, categories["Женская одежда"].ID, units["шт"].ID, "Размер", []string{"XS", "S", "M"})
	s.generateVariants(&allVariants, "Кроссовки", []string{"New Balance 574", "Reebok Classic"}, categories["Мужская обувь"].ID, units["пара"].ID, "Размер", []string{"42", "43", "44"})
	s.generateVariants(&allVariants, "Туфли", []string{"ECCO Helsinki", "Geox Wells"}, categories["Женская обувь"].ID, units["пара"].ID, "Размер", []string{"37", "38", "39"})

	// --- СПОРТ И ХОББИ ---
	s.generateSimpleProducts(&allVariants, []string{"Палатка 3-местная", "Спальный мешок -10С", "Рюкзак туристический 80л"}, categories["Товары для туризма"].ID, units["шт"].ID)
	s.generateSimpleProducts(&allVariants, []string{"Набор гантелей", "Коврик для йоги", "Футбольный мяч"}, categories["Спортивные товары"].ID, units["шт"].ID)
	s.generateSimpleProducts(&allVariants, []string{"Война и мир", "Мастер и Маргарита"}, categories["Книги"].ID, units["шт"].ID)

	// --- ПРОЧЕЕ ---
	s.generateSimpleProducts(&allVariants, []string{"Набор инструментов 100 предметов", "Шуруповерт аккумуляторный"}, categories["Инструменты"].ID, units["шт"].ID)
	s.generateSimpleProducts(&allVariants, []string{"Корм для собак, 15кг", "Наполнитель для кошачьего туалета, 10л"}, categories["Товары для животных"].ID, units["упак"].ID)

	log.Printf("Сгенерировано %d вариантов товаров.", len(allVariants))
	return allVariants
}

// generateVariants - вспомогательная функция для генерации продуктов с вариантами
func (s *seeder) generateVariants(
	allVariants *[]models.Variant,
	baseName string, productNames []string, categoryID uint, unitID uint,
	charName string, charValues []string,
) {
	for _, pName := range productNames {
		p, _ := s.productSvc.Create(&models.Product{Name: pName, CategoryID: categoryID})
		for _, cValue := range charValues {
			sku := fmt.Sprintf("%s-%s-%s", strings.ToUpper(pName[:3]), strings.ToUpper(baseName[:3]), strings.ReplaceAll(cValue, "/", ""))
			v, _ := s.variantSvc.Create(&models.Variant{
				ProductID: p.ID, SKU: sku, UnitID: unitID,
				Characteristics: models.CharacteristicsMap{charName: cValue},
			})
			*allVariants = append(*allVariants, *v)
		}
	}
}

// generateSimpleProducts - вспомогательная функция для товаров без вариантов
func (s *seeder) generateSimpleProducts(
	allVariants *[]models.Variant,
	productNames []string, categoryID uint, unitID uint,
) {
	for _, pName := range productNames {
		p, _ := s.productSvc.Create(&models.Product{Name: pName, CategoryID: categoryID})
		sku := fmt.Sprintf("%s-%d", strings.ToUpper(pName[:4]), rand.Intn(1000))
		v, _ := s.variantSvc.Create(&models.Variant{ProductID: p.ID, SKU: sku, UnitID: unitID})
		*allVariants = append(*allVariants, *v)
	}
}

// createIncomeDocument теперь распределяет товары по ДВУМ складам
func (s *seeder) createIncomeDocument(warehouses []models.Warehouse, variants []models.Variant, units map[string]models.Unit) {
	if len(warehouses) < 2 {
		log.Println("Нужно как минимум 2 склада для распределения, приходуем все на первый.")
		warehouses = append(warehouses, warehouses[0])
	}

	itemsForWh1 := []models.DocumentItem{}
	itemsForWh2 := []models.DocumentItem{}

	for i, v := range variants {
		var qty decimal.Decimal
		if v.UnitID == units["кг"].ID {
			qty = decimal.NewFromFloat(5 + rand.Float64()*(50-5)).Round(2)
		} else {
			qty = decimal.NewFromInt(int64(10 + rand.Intn(91)))
		}
		price, _ := decimal.NewFromString(fmt.Sprintf("%d", 500+rand.Intn(15000)))

		item := models.DocumentItem{VariantID: v.ID, Quantity: qty, Price: &price}

		// Распределяем товары по складам (четные/нечетные)
		if i%2 == 0 {
			itemsForWh1 = append(itemsForWh1, item)
		} else {
			itemsForWh2 = append(itemsForWh2, item)
		}
	}

	s.postIncomeDoc(warehouses[0].ID, "Начальное наполнение склада (Москва)", itemsForWh1)
	s.postIncomeDoc(warehouses[1].ID, "Начальное наполнение склада (СПб)", itemsForWh2)
}

// Вспомогательная функция для создания и проведения одного документа прихода
func (s *seeder) postIncomeDoc(whID uint, comment string, items []models.DocumentItem) {
	if len(items) == 0 {
		return
	}

	doc, err := s.docSvc.Create(&models.Document{
		Type:        "INCOME",
		WarehouseID: &whID,
		CreatedAt:   time.Now(),
		Comment:     comment,
		Items:       items,
	})
	if err != nil {
		log.Fatalf("Не удалось создать документ прихода: %v", err)
	}

	err = s.docSvc.Post(doc.ID)
	if err != nil {
		log.Fatalf("Не удалось провести документ прихода: %v", err)
	}
	log.Printf("Проведен документ прихода '%s' с %d позициями на склад ID %d.", doc.Number, len(items), whID)
}
