package httpserver

import "hexagon/contact"

type AddContactRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (r AddContactRequest) ToContact() contact.Contact {
	return contact.Contact{
		Name:  r.Name,
		Phone: r.Phone,
	}
}
