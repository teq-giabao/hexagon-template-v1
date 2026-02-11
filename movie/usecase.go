package movie

import "context"

type Service interface {
	Search(ctx context.Context, query string, limit int) ([]Movie, error)
}

type Repository interface {
	Search(ctx context.Context, query string, limit int) ([]Movie, error)
}

type Usecase struct {
	r Repository
}

func NewUsecase(r Repository) *Usecase {
	return &Usecase{r: r}
}

func (uc *Usecase) Search(ctx context.Context, query string, limit int) ([]Movie, error) {
	return uc.r.Search(ctx, query, limit)
}
