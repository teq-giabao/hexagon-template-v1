package movie

import "hexagon/errs"

var ErrInvalidQuery = errs.Errorf(errs.EINVALID, "invalid search query")

type Movie struct {
	MovieID int    `json:"movieId"`
	Title   string `json:"title"`
	Genres  string `json:"genres"`
}
