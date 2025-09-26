package service

import (
	stock "github.com/maksroxx/flowkeeper/internal/modules/stock/models"
	"github.com/maksroxx/flowkeeper/internal/modules/stock/repository"
)

type UnitService interface {
	Create(name string) (*stock.Unit, error)
	GetByID(id uint) (*stock.Unit, error)
	List() ([]stock.Unit, error)
	Update(u *stock.Unit) (*stock.Unit, error)
	Delete(id uint) error
}

type unitService struct{ repo repository.UnitRepository }

func NewUnitService(r repository.UnitRepository) UnitService { return &unitService{repo: r} }

func (s *unitService) Create(name string) (*stock.Unit, error) {
	return s.repo.Create(&stock.Unit{Name: name})
}
func (s *unitService) GetByID(id uint) (*stock.Unit, error)      { return s.repo.GetByID(id) }
func (s *unitService) List() ([]stock.Unit, error)               { return s.repo.List() }
func (s *unitService) Update(u *stock.Unit) (*stock.Unit, error) { return s.repo.Update(u) }
func (s *unitService) Delete(id uint) error                      { return s.repo.Delete(id) }
