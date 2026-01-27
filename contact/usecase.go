package contact

import "context"

type Service interface {
	AddContact(ctx context.Context, c Contact) error
	ListContacts(ctx context.Context) ([]Contact, error)
}

type Repository interface {
	CreateContact(ctx context.Context, c Contact) error
	AllContacts(ctx context.Context) ([]Contact, error)
}

type Usecase struct {
	r Repository
}

func NewUsecase(r Repository) *Usecase {
	return &Usecase{r: r}
}

func (uc *Usecase) AddContact(ctx context.Context, c Contact) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return uc.r.CreateContact(ctx, c)
}

func (uc *Usecase) ListContacts(ctx context.Context) ([]Contact, error) {
	return uc.r.AllContacts(ctx)
}
