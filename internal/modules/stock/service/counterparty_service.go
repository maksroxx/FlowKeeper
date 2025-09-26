package service

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type CounterpartyService interface {
	Create(name string) (*stock.Counterparty, error)
	GetByID(id uint) (*stock.Counterparty, error)
	List() ([]stock.Counterparty, error)
	Update(cp *stock.Counterparty) (*stock.Counterparty, error)
	Delete(id uint) error
}

type counterpartyService struct {
	repo repository.CounterpartyRepository
}

func NewCounterpartyService(r repository.CounterpartyRepository) CounterpartyService {
	return &counterpartyService{repo: r}
}

func (s *counterpartyService) Create(name string) (*stock.Counterparty, error) {
	return s.repo.Create(&stock.Counterparty{Name: name})
}
func (s *counterpartyService) GetByID(id uint) (*stock.Counterparty, error) {
	return s.repo.GetByID(id)
}
func (s *counterpartyService) List() ([]stock.Counterparty, error) { return s.repo.List() }
func (s *counterpartyService) Update(cp *stock.Counterparty) (*stock.Counterparty, error) {
	return s.repo.Update(cp)
}
func (s *counterpartyService) Delete(id uint) error { return s.repo.Delete(id) }
