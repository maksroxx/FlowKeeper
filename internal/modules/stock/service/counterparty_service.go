package service

import (
	"errors"
	"net/mail"
	"strings"

	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type CounterpartyService interface {
	Create(cp *stock.Counterparty) (*stock.Counterparty, error)
	GetByID(id uint) (*stock.Counterparty, error)
	List() ([]stock.Counterparty, error)
	Update(id uint, updates map[string]interface{}) (*stock.Counterparty, error)
	Delete(id uint) error

	Search(filter stock.CounterpartyFilter) ([]stock.Counterparty, error)
}

type counterpartyService struct {
	repo repository.CounterpartyRepository
}

func NewCounterpartyService(r repository.CounterpartyRepository) CounterpartyService {
	return &counterpartyService{repo: r}
}

func (s *counterpartyService) Create(cp *stock.Counterparty) (*stock.Counterparty, error) {
	return s.repo.Create(cp)
}
func (s *counterpartyService) GetByID(id uint) (*stock.Counterparty, error) {
	return s.repo.GetByID(id)
}
func (s *counterpartyService) List() ([]stock.Counterparty, error) { return s.repo.List() }
func (s *counterpartyService) Update(id uint, updates map[string]interface{}) (*stock.Counterparty, error) {
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	delete(updates, "deleted_at")

	for key, value := range updates {
		switch strings.ToLower(key) {

		case "name":
			name, ok := value.(string)
			if !ok {
				return nil, errors.New("invalid type for 'name', expected string")
			}
			if strings.TrimSpace(name) == "" {
				return nil, errors.New("'name' cannot be empty")
			}

		case "email":
			email, ok := value.(string)
			if !ok {
				return nil, errors.New("invalid type for 'email', expected string")
			}
			if email != "" {
				if _, err := mail.ParseAddress(email); err != nil {
					return nil, errors.New("invalid email format")
				}
			}

		case "phone":
			if _, ok := value.(string); !ok {
				return nil, errors.New("invalid type for 'phone', expected string")
			}
		}
	}

	if len(updates) == 0 {
		return s.repo.GetByID(id)
	}

	return s.repo.Patch(id, updates)
}
func (s *counterpartyService) Delete(id uint) error { return s.repo.Delete(id) }

func (s *counterpartyService) Search(filter stock.CounterpartyFilter) ([]stock.Counterparty, error) {
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	return s.repo.Search(filter)
}
