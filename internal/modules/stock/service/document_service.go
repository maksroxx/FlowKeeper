package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type DocumentService interface {
	Post(id uint) error
	Cancel(id uint) error
	Create(doc *models.Document) (*models.Document, error)
	GetByID(id uint) (*models.Document, error)
	GetByIDAsDTO(id uint) (*models.DocumentDTO, error)
	ListAsDTO(status string) ([]models.DocumentListItemDTO, error)
	Update(id uint, updatePayload *models.DocumentUpdateDTO) (*models.DocumentDTO, error)
	Delete(id uint) error

	SearchAsDTO(filter models.DocumentFilter) ([]models.DocumentListItemDTO, error)
}

type documentService struct {
	repo         repository.DocumentRepository
	historyRepo  repository.DocumentHistoryRepository
	inventory    InventoryService
	priceService PriceService
	sequenceSvc  SequenceService
	tx           repository.TxManager
	variantRepo  repository.VariantRepository
	productRepo  repository.ProductRepository
	whRepo       repository.WarehouseRepository
	cpRepo       repository.CounterpartyRepository
	ptRepo       repository.PriceTypeRepository
}

func NewDocumentService(
	repo repository.DocumentRepository, historyRepo repository.DocumentHistoryRepository, inventory InventoryService,
	priceService PriceService, sequenceSvc SequenceService, tx repository.TxManager,
	variantRepo repository.VariantRepository, productRepo repository.ProductRepository, whRepo repository.WarehouseRepository,
	cpRepo repository.CounterpartyRepository, ptRepo repository.PriceTypeRepository,
) DocumentService {
	return &documentService{
		repo: repo, historyRepo: historyRepo, inventory: inventory, priceService: priceService,
		sequenceSvc: sequenceSvc, tx: tx, variantRepo: variantRepo, productRepo: productRepo,
		whRepo: whRepo, cpRepo: cpRepo, ptRepo: ptRepo,
	}
}

func (s *documentService) Post(id uint) error {
	err := s.tx.DoInTx(func(tx *gorm.DB) error {
		doc, err := s.repo.GetByIDWithTx(tx, id)
		if err != nil {
			return err
		}
		if doc == nil {
			return errors.New("document not found")
		}
		if doc.Status == "posted" {
			return errors.New("document already posted")
		}
		if doc.Status == "canceled" {
			return errors.New("cannot post canceled document")
		}

		switch toUpper(doc.Type) {
		case "INCOME", "OUTCOME", "ORDER", "TRANSFER", "INVENTORY":
			if err := s.inventory.ProcessDocumentWithTx(tx, doc); err != nil {
				return fmt.Errorf("inventory processing failed: %w", err)
			}
		case "PRICE_UPDATE":
			if err := s.priceService.UpdatePricesFromDocumentWithTx(tx, doc); err != nil {
				return fmt.Errorf("price update processing failed: %w", err)
			}
		default:
			return fmt.Errorf("unknown document type to post: '%s'", doc.Type)
		}

		now := time.Now()
		doc.Status = "posted"
		doc.PostedAt = &now
		if _, err := s.repo.UpdateWithTx(tx, doc); err != nil {
			return err
		}

		h := &models.DocumentHistory{DocumentID: doc.ID, Action: "posted", CreatedAt: now, CreatedBy: doc.CreatedBy}
		return s.historyRepo.CreateWithTx(tx, h)
	})
	if err != nil {
		log.Printf("[ERROR] Failed to post document ID=%d: %v", id, err)
		return err
	}
	return nil
}

func (s *documentService) Cancel(id uint) error {
	return s.tx.DoInTx(func(tx *gorm.DB) error {
		doc, err := s.repo.GetByIDWithTx(tx, id)
		if err != nil {
			return err
		}
		if doc == nil {
			return errors.New("document not found")
		}
		if doc.Status != "posted" {
			return errors.New("only posted documents can be canceled")
		}

		switch toUpper(doc.Type) {
		case "INCOME", "OUTCOME", "ORDER", "TRANSFER", "INVENTORY":
			if err := s.inventory.RevertDocumentWithTx(tx, doc); err != nil {
				return err
			}
		case "PRICE_UPDATE":
			return errors.New("cancellation for 'PRICE_UPDATE' is not yet implemented")
		default:
			return fmt.Errorf("unknown document type to cancel: '%s'", doc.Type)
		}

		doc.Status = "canceled"
		if _, err := s.repo.UpdateWithTx(tx, doc); err != nil {
			return err
		}

		h := &models.DocumentHistory{DocumentID: doc.ID, Action: "canceled", CreatedAt: time.Now(), CreatedBy: doc.CreatedBy}
		return s.historyRepo.CreateWithTx(tx, h)
	})
}

func (s *documentService) Create(doc *models.Document) (*models.Document, error) {
	if doc.Status == "" {
		doc.Status = "draft"
	}
	newNumber, err := s.sequenceSvc.GenerateNextDocumentNumber(doc.Type)
	if err != nil {
		return nil, fmt.Errorf("could not generate document number: %w", err)
	}
	doc.Number = newNumber
	return s.repo.Create(doc)
}

func (s *documentService) GetByID(id uint) (*models.Document, error) { return s.repo.GetByID(id) }

func (s *documentService) Update(id uint, updatePayload *models.DocumentUpdateDTO) (*models.DocumentDTO, error) {
	var finalDoc *models.Document

	err := s.tx.DoInTx(func(tx *gorm.DB) error {
		docToUpdate, err := s.repo.GetByIDWithTx(tx, id)
		if err != nil {
			return err
		}
		if docToUpdate == nil {
			return errors.New("document not found")
		}

		if docToUpdate.Status != "draft" {
			return errors.New("only draft documents can be edited")
		}

		docToUpdate.WarehouseID = updatePayload.WarehouseID
		docToUpdate.CounterpartyID = updatePayload.CounterpartyID
		docToUpdate.Comment = updatePayload.Comment

		if err := tx.Where("document_id = ?", docToUpdate.ID).Delete(&models.DocumentItem{}).Error; err != nil {
			return fmt.Errorf("failed to delete old document items: %w", err)
		}

		docToUpdate.Items = updatePayload.Items

		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(docToUpdate).Error; err != nil {
			return fmt.Errorf("failed to save updated document: %w", err)
		}

		finalDoc = docToUpdate
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildDTO(finalDoc)
}

func (s *documentService) Delete(id uint) error { return s.repo.Delete(id) }

func (s *documentService) GetByIDAsDTO(id uint) (*models.DocumentDTO, error) {
	doc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	return s.buildDTO(doc)
}

func (s *documentService) ListAsDTO(status string) ([]models.DocumentListItemDTO, error) {
	var docs []models.Document
	var err error
	if status != "" {
		docs, err = s.repo.ListByStatus(status)
	} else {
		docs, err = s.repo.List()
	}
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return []models.DocumentListItemDTO{}, nil
	}

	warehouseIDs := make(map[uint]bool)
	counterpartyIDs := make(map[uint]bool)
	for _, doc := range docs {
		if doc.WarehouseID != nil {
			warehouseIDs[*doc.WarehouseID] = true
		}
		if doc.CounterpartyID != nil {
			counterpartyIDs[*doc.CounterpartyID] = true
		}
	}
	whMap, _ := s.loadWarehouseMap(mapKeysToSlice(warehouseIDs))
	cpMap, _ := s.loadCounterpartyMap(mapKeysToSlice(counterpartyIDs))

	dtos := make([]models.DocumentListItemDTO, len(docs))
	for i, doc := range docs {
		var whName, cpName string
		if doc.WarehouseID != nil {
			whName = whMap[*doc.WarehouseID]
		}
		if doc.CounterpartyID != nil {
			cpName = cpMap[*doc.CounterpartyID]
		}

		dtos[i] = models.DocumentListItemDTO{
			ID:               doc.ID,
			Type:             doc.Type,
			Number:           doc.Number,
			WarehouseName:    whName,
			CounterpartyName: cpName,
			ItemCount:        len(doc.Items),
			Status:           doc.Status,
			CreatedAt:        doc.CreatedAt,
		}
	}

	return dtos, nil
}

func (s *documentService) SearchAsDTO(filter models.DocumentFilter) ([]models.DocumentListItemDTO, error) {
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	docs, err := s.repo.Search(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	if len(docs) == 0 {
		return []models.DocumentListItemDTO{}, nil
	}

	warehouseIDs := make(map[uint]bool)
	counterpartyIDs := make(map[uint]bool)

	for _, doc := range docs {
		if doc.WarehouseID != nil {
			warehouseIDs[*doc.WarehouseID] = true
		}
		if doc.CounterpartyID != nil {
			counterpartyIDs[*doc.CounterpartyID] = true
		}
	}

	whMap, err := s.loadWarehouseMap(mapKeysToSlice(warehouseIDs))
	if err != nil {
		return nil, err
	}

	cpMap, err := s.loadCounterpartyMap(mapKeysToSlice(counterpartyIDs))
	if err != nil {
		return nil, err
	}

	dtos := make([]models.DocumentListItemDTO, len(docs))
	for i, doc := range docs {
		var whName, cpName string
		if doc.WarehouseID != nil {
			whName = whMap[*doc.WarehouseID]
		}
		if doc.CounterpartyID != nil {
			cpName = cpMap[*doc.CounterpartyID]
		}

		dtos[i] = models.DocumentListItemDTO{
			ID:               doc.ID,
			Type:             doc.Type,
			Number:           doc.Number,
			WarehouseName:    whName,
			CounterpartyName: cpName,
			ItemCount:        len(doc.Items),
			Status:           doc.Status,
			CreatedAt:        doc.CreatedAt,
		}
	}

	return dtos, nil
}

func (s *documentService) buildDTO(doc *models.Document) (*models.DocumentDTO, error) {
	dto := &models.DocumentDTO{
		ID: doc.ID, Type: doc.Type, Number: doc.Number, Comment: doc.Comment,
		BaseDocumentID: doc.BaseDocumentID, Status: doc.Status, PostedAt: doc.PostedAt, CreatedAt: doc.CreatedAt,
		WarehouseID: doc.WarehouseID, ToWarehouseID: doc.ToWarehouseID, CounterpartyID: doc.CounterpartyID, PriceTypeID: doc.PriceTypeID,
	}

	if len(doc.Items) > 0 {
		variantIDs := make([]uint, len(doc.Items))
		for i, item := range doc.Items {
			variantIDs[i] = item.VariantID
		}

		variantMap, productMap := s.loadVariantAndProductMaps(variantIDs)

		itemDTOs := make([]models.DocumentItemDTO, len(doc.Items))
		for i, item := range doc.Items {
			variant := variantMap[item.VariantID]
			product := productMap[variant.ProductID]
			itemDTOs[i] = models.DocumentItemDTO{
				ID: item.ID, VariantID: item.VariantID, VariantSKU: variant.SKU,
				ProductName: product.Name, Quantity: item.Quantity, Price: item.Price,
			}
		}
		dto.Items = itemDTOs
	}

	if doc.WarehouseID != nil {
		if wh, _ := s.whRepo.GetByID(*doc.WarehouseID); wh != nil {
			dto.WarehouseName = wh.Name
		}
	}
	if doc.ToWarehouseID != nil {
		if wh, _ := s.whRepo.GetByID(*doc.ToWarehouseID); wh != nil {
			dto.ToWarehouseName = wh.Name
		}
	}
	if doc.CounterpartyID != nil {
		if cp, _ := s.cpRepo.GetByID(*doc.CounterpartyID); cp != nil {
			dto.CounterpartyName = cp.Name
		}
	}
	if doc.PriceTypeID != nil {
		if pt, _ := s.ptRepo.GetByID(*doc.PriceTypeID); pt != nil {
			dto.PriceTypeName = pt.Name
		}
	}
	return dto, nil
}

func (s *documentService) loadVariantAndProductMaps(ids []uint) (map[uint]models.Variant, map[uint]models.Product) {
	variants, _ := s.variantRepo.GetByIDs(ids)
	variantMap := make(map[uint]models.Variant, len(variants))
	productIDs := make(map[uint]bool)
	for _, v := range variants {
		variantMap[v.ID] = v
		if v.ProductID != 0 {
			productIDs[v.ProductID] = true
		}
	}
	products, _ := s.productRepo.GetByIDs(mapKeysToSlice(productIDs))
	productMap := make(map[uint]models.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}
	return variantMap, productMap
}

func (s *documentService) loadWarehouseMap(ids []uint) (map[uint]string, error) {
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

func (s *documentService) loadCounterpartyMap(ids []uint) (map[uint]string, error) {
	counterparties, err := s.cpRepo.GetByIDs(ids)
	if err != nil {
		return nil, err
	}
	cpMap := make(map[uint]string, len(counterparties))
	for _, cp := range counterparties {
		cpMap[cp.ID] = cp.Name
	}
	return cpMap, nil
}

func mapKeysToSlice(m map[uint]bool) []uint {
	s := make([]uint, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}
