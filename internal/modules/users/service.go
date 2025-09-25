package users

type UserService interface {
	AddUser(name, email, password, role string) (*User, error)
	ListUsers() ([]User, error)
	GetUser(id uint) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id uint) error
}

type userService struct {
	repo UserRepository
}

func NewService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) AddUser(name, email, password, role string) (*User, error) {
	user := &User{Name: name, Email: email, Password: password, Role: role}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ListUsers() ([]User, error) {
	return s.repo.GetAll()
}

func (s *userService) GetUser(id uint) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) UpdateUser(user *User) error {
	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}
