package stock

type StockService interface {
	AddItem(name string, stock int) (*Item, error)
	ListItems() ([]Item, error)
}

type stockService struct {
	repo ItemRepository
}

func (s *stockService) AddItem(name string, stock int) (*Item, error) {
	item := &Item{Name: name, Stock: stock}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *stockService) ListItems() ([]Item, error) {
	return s.repo.List()
}
