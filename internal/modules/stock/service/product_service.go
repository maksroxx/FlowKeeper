package service

import (
	"encoding/csv"
	"errors"
	"mime/multipart"
	"strings"

	"github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
	"github.com/xuri/excelize/v2"
)

type ProductService interface {
	Create(p *models.Product, sku string, unitID uint, characteristics map[string]string, images []string) (*models.Product, error)
	GetByID(id uint) (*models.Product, error)
	List() ([]models.Product, error)
	Update(id uint, updates map[string]interface{}) (*models.Product, error)
	Delete(id uint) error

	GetProductDetails(productID uint) (*models.ProductDetailDTO, error)
	ImportItems(file multipart.File, filename string) error
}

type productService struct {
	repo        repository.ProductRepository
	variantRepo repository.VariantRepository
}

func NewProductService(repo repository.ProductRepository, variantRepo repository.VariantRepository) ProductService {
	return &productService{repo: repo, variantRepo: variantRepo}
}

func (s *productService) Create(p *models.Product, sku string, unitID uint, char map[string]string, images []string) (*models.Product, error) {
	createdProduct, err := s.repo.Create(p)
	if err != nil {
		return nil, err
	}

	v := &models.Variant{
		ProductID:       createdProduct.ID,
		SKU:             sku,
		UnitID:          unitID,
		Characteristics: char,
	}

	for _, url := range images {
		v.Images = append(v.Images, models.ProductImage{URL: url})
	}

	if _, err := s.variantRepo.Create(v); err != nil {
		return nil, err
	}

	return createdProduct, nil
}

func (s *productService) GetByID(id uint) (*models.Product, error) {
	return s.repo.GetByID(id)
}

func (s *productService) List() ([]models.Product, error) {
	return s.repo.List()
}

// func (s *productService) Update(id uint, updateData *models.Product) (*models.Product, error) {
// 	productToUpdate, err := s.repo.GetByID(id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if productToUpdate == nil {
// 		return nil, errors.New("product not found to update")
// 	}

// 	productToUpdate.Name = updateData.Name
// 	productToUpdate.Description = updateData.Description
// 	productToUpdate.CategoryID = updateData.CategoryID

// 	return s.repo.Update(productToUpdate)
// }

func (s *productService) Update(id uint, updates map[string]interface{}) (*models.Product, error) {
	return s.repo.Patch(id, updates)
}

func (s *productService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *productService) GetProductDetails(productID uint) (*models.ProductDetailDTO, error) {
	product, err := s.repo.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	variants, err := s.variantRepo.FindByProductID(productID)
	if err != nil {
		return nil, err
	}

	optionsMap := make(map[string]map[string]bool)
	for _, v := range variants {
		for charType, charValue := range v.Characteristics {
			if _, ok := optionsMap[charType]; !ok {
				optionsMap[charType] = make(map[string]bool)
			}
			optionsMap[charType][charValue] = true
		}
	}

	options := make([]models.ProductOptionDTO, 0, len(optionsMap))
	for charType, valuesMap := range optionsMap {
		values := make([]string, 0, len(valuesMap))
		for value := range valuesMap {
			values = append(values, value)
		}
		options = append(options, models.ProductOptionDTO{Type: charType, Values: values})
	}

	details := &models.ProductDetailDTO{
		Product:  *product,
		Variants: variants,
		Options:  options,
	}

	return details, nil
}

func (s *productService) ImportItems(file multipart.File, filename string) error {
	var rows [][]string
	var err error

	if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		f, err := excelize.OpenReader(file)
		if err != nil {
			return err
		}
		sheetName := f.GetSheetName(0)
		if sheetName == "" {
			return errors.New("xlsx file is empty")
		}
		rows, err = f.GetRows(sheetName)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		reader := csv.NewReader(file)
		reader.Comma = ';'
		reader.LazyQuotes = true
		rows, err = reader.ReadAll()
		if err != nil {
			return err
		}
	} else {
		return errors.New("unsupported file format (use .csv or .xlsx)")
	}

	if len(rows) > 0 {
		firstCell := strings.ToLower(strings.TrimSpace(rows[0][0]))
		if strings.Contains(firstCell, "категория") || strings.Contains(firstCell, "category") {
			rows = rows[1:]
		}
	}

	var importData []models.ImportItemDTO

	for _, row := range rows {
		if len(row) < 3 {
			continue
		}

		catName := strings.TrimSpace(row[0])
		prodName := strings.TrimSpace(row[1])
		sku := strings.TrimSpace(row[2])

		if prodName == "" || sku == "" {
			continue
		}
		if catName == "" {
			catName = "Общее"
		}

		unitName := "шт"
		if len(row) > 3 && row[3] != "" {
			unitName = strings.TrimSpace(row[3])
		}

		chars := make(map[string]string)
		if len(row) > 4 && row[4] != "" {
			parts := strings.Split(row[4], ";")
			for _, part := range parts {
				kv := strings.Split(part, ":")
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					val := strings.TrimSpace(kv[1])
					if key != "" && val != "" {
						chars[key] = val
					}
				}
			}
		}

		desc := ""
		if len(row) > 5 {
			desc = strings.TrimSpace(row[5])
		}

		importData = append(importData, models.ImportItemDTO{
			CategoryName:    catName,
			ProductName:     prodName,
			SKU:             sku,
			UnitName:        unitName,
			Characteristics: chars,
			Description:     desc,
		})
	}

	if len(importData) == 0 {
		return errors.New("no valid data found in file")
	}

	return s.repo.ImportBatch(importData)
}
