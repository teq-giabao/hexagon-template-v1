package httpserver

import (
	"hexagon/errs"
	"hexagon/movie"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterPublicMovieRoutes(g *echo.Group) {
	g.GET("/movies/search", s.handleSearchMovies)
}

// handleSearchMovies godoc
// @Summary Search Movies
// @Description Full-text search movies by title/genres
// @Tags movies
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Max results (1-100), default 20"
// @Success 200 {array} movie.Movie
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/movies/search [get]
func (s *Server) handleSearchMovies(c echo.Context) error {
	if s.MovieService == nil {
		return errs.Errorf(errs.ENOTIMPLEMENTED, "movie service not configured")
	}

	query := strings.TrimSpace(c.QueryParam("q"))
	if query == "" {
		return movie.ErrInvalidQuery
	}

	limit := 20
	if raw := strings.TrimSpace(c.QueryParam("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return movie.ErrInvalidQuery
		}
		limit = parsed
	}

	results, err := s.MovieService.Search(c.Request().Context(), query, limit)
	if err != nil {
		return err
	}

	return RespondList(c, http.StatusOK, results)
}
