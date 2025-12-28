package click

type Service interface {
	AddClick(urlID int64)
	GetClicks(urlID int64) (int, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) AddClick(urlID int64) {
	_ = s.repo.Add(urlID)
}

func (s *service) GetClicks(urlID int64) (int, error) {
	return s.repo.Count(urlID)
}
