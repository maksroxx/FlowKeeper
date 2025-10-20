package service

import (
	"time"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/shopspring/decimal"
)

type StockMovementService interface {
	Create(itemID, warehouseID uint, counterpartyID *uint, qty decimal.Decimal, mtype, comment string) (*stock.StockMovement, error)
	GetByID(id uint) (*stock.StockMovement, error)
	List() ([]stock.StockMovement, error)
	Update(m *stock.StockMovement) (*stock.StockMovement, error)
	Delete(id uint) error
	ListAsDTO() ([]stock.StockMovementDTO, error)
	SearchAsDTO(filter stock.MovementFilter) ([]stock.StockMovementDTO, error)
}

type movementService struct {
	repo        repository.StockMovementRepository
	docRepo     repository.DocumentRepository
	variantRepo repository.VariantRepository
	productRepo repository.ProductRepository
	whRepo      repository.WarehouseRepository
}

func NewStockMovementService(r repository.StockMovementRepository, docRepo repository.DocumentRepository,
	variantRepo repository.VariantRepository,
	productRepo repository.ProductRepository,
	whRepo repository.WarehouseRepository) StockMovementService {
	return &movementService{repo: r, docRepo: docRepo,
		variantRepo: variantRepo,
		productRepo: productRepo,
		whRepo:      whRepo}
}

func (s *movementService) Create(itemID, warehouseID uint, counterpartyID *uint, qty decimal.Decimal, mtype, comment string) (*stock.StockMovement, error) {
	return s.repo.Create(&stock.StockMovement{
		VariantID:      itemID,
		WarehouseID:    warehouseID,
		CounterpartyID: counterpartyID,
		Quantity:       qty,
		Type:           mtype,
		Comment:        comment,
		CreatedAt:      time.Now(),
	})
}
func (s *movementService) GetByID(id uint) (*stock.StockMovement, error) { return s.repo.GetByID(id) }
func (s *movementService) List() ([]stock.StockMovement, error)          { return s.repo.List() }
func (s *movementService) Update(m *stock.StockMovement) (*stock.StockMovement, error) {
	return s.repo.Update(m)
}
func (s *movementService) Delete(id uint) error { return s.repo.Delete(id) }

// deprecated
func (s *movementService) ListAsDTO() ([]stock.StockMovementDTO, error) {
	movements, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	if len(movements) == 0 {
		return []stock.StockMovementDTO{}, nil
	}

	docIDs := make(map[uint]bool)
	variantIDs := make(map[uint]bool)
	warehouseIDs := make(map[uint]bool)

	for _, mv := range movements {
		if mv.DocumentID != nil {
			docIDs[*mv.DocumentID] = true
		}
		variantIDs[mv.VariantID] = true
		warehouseIDs[mv.WarehouseID] = true
	}

	documents, _ := s.docRepo.GetByIDs(mapKeysToSlice(docIDs))
	docMap := make(map[uint]string)
	for _, d := range documents {
		docMap[d.ID] = d.Number
	}

	variants, _ := s.variantRepo.GetByIDs(mapKeysToSlice(variantIDs))
	variantMap := make(map[uint]stock.Variant)
	productIDsMap := make(map[uint]bool)
	for _, v := range variants {
		variantMap[v.ID] = v
		productIDsMap[v.ProductID] = true
	}

	products, _ := s.productRepo.GetByIDs(mapKeysToSlice(productIDsMap))
	productMap := make(map[uint]string)
	for _, p := range products {
		productMap[p.ID] = p.Name
	}

	warehouses, _ := s.whRepo.GetByIDs(mapKeysToSlice(warehouseIDs))
	whMap := make(map[uint]string)
	for _, wh := range warehouses {
		whMap[wh.ID] = wh.Name
	}

	dtos := make([]stock.StockMovementDTO, len(movements))
	for i, mv := range movements {
		variant := variantMap[mv.VariantID]
		productName := productMap[variant.ProductID]

		dtos[i] = stock.StockMovementDTO{
			ID:             mv.ID,
			DocumentID:     mv.DocumentID,
			DocumentNumber: docMap[*mv.DocumentID],
			VariantID:      mv.VariantID,
			VariantSKU:     variant.SKU,
			ProductName:    productName,
			WarehouseID:    mv.WarehouseID,
			WarehouseName:  whMap[mv.WarehouseID],
			Quantity:       mv.Quantity,
			Type:           mv.Type,
			CreatedAt:      mv.CreatedAt,
		}
	}

	return dtos, nil
}

func (s *movementService) SearchAsDTO(filter stock.MovementFilter) ([]stock.StockMovementDTO, error) {
	movements, err := s.repo.Search(filter)
	if err != nil {
		return nil, err
	}
	if len(movements) == 0 {
		return []stock.StockMovementDTO{}, nil
	}

	docIDs := make(map[uint]bool)
	variantIDs := make(map[uint]bool)
	warehouseIDs := make(map[uint]bool)
	for _, mv := range movements {
		if mv.DocumentID != nil {
			docIDs[*mv.DocumentID] = true
		}
		variantIDs[mv.VariantID] = true
		warehouseIDs[mv.WarehouseID] = true
	}

	docMap, _ := s.loadDocumentMap(mapKeysToSlice(docIDs))
	variantMap, productMap := s.loadVariantAndProductMaps(mapKeysToSlice(variantIDs))
	whMap, _ := s.loadWarehouseMap(mapKeysToSlice(warehouseIDs))

	dtos := make([]stock.StockMovementDTO, len(movements))
	for i, mv := range movements {
		variant := variantMap[mv.VariantID]
		productName := productMap[variant.ProductID]

		var docNum string
		if mv.DocumentID != nil {
			docNum = docMap[*mv.DocumentID]
		}

		dtos[i] = stock.StockMovementDTO{
			ID:             mv.ID,
			DocumentID:     mv.DocumentID,
			DocumentNumber: docNum,
			VariantID:      mv.VariantID,
			VariantSKU:     variant.SKU,
			ProductName:    productName,
			WarehouseID:    mv.WarehouseID,
			WarehouseName:  whMap[mv.WarehouseID],
			Quantity:       mv.Quantity,
			Type:           mv.Type,
			CreatedAt:      mv.CreatedAt,
		}
	}
	return dtos, nil
}

func (s *movementService) loadDocumentMap(ids []uint) (map[uint]string, error) {
	documents, err := s.docRepo.GetByIDs(ids)
	if err != nil {
		return nil, err
	}
	docMap := make(map[uint]string, len(documents))
	for _, d := range documents {
		docMap[d.ID] = d.Number
	}
	return docMap, nil
}
func (s *movementService) loadVariantAndProductMaps(ids []uint) (map[uint]stock.Variant, map[uint]string) {
	variants, _ := s.variantRepo.GetByIDs(ids)
	variantMap := make(map[uint]stock.Variant, len(variants))
	productIDs := make(map[uint]bool)
	for _, v := range variants {
		variantMap[v.ID] = v
		if v.ProductID != 0 {
			productIDs[v.ProductID] = true
		}
	}
	products, _ := s.productRepo.GetByIDs(mapKeysToSlice(productIDs))
	productMap := make(map[uint]string, len(products))
	for _, p := range products {
		productMap[p.ID] = p.Name
	}
	return variantMap, productMap
}
func (s *movementService) loadWarehouseMap(ids []uint) (map[uint]string, error) {
	warehouses, err := s.whRepo.GetByIDs(ids)
	if err != nil {
		return nil, err
	}
	whMap := make(map[uint]string, len(warehouses))
	for _, wh := range warehouses {
		whMap[wh.ID] = wh.Name
	}
	return whMap, nil
}
