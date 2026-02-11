package postgres

import (
	"context"
	"hexagon/movie"

	"gorm.io/gorm"
)

// MovieModel represents the database model for movies
// search_vector is generated in SQL migration and not mapped here.
type MovieModel struct {
	ID      uint   `gorm:"primaryKey"`
	MovieID int    `gorm:"column:movie_id;not null;uniqueIndex"`
	Title   string `gorm:"not null"`
	Genres  string `gorm:"not null;default:''"`
}

// TableName specifies the table name for GORM
func (MovieModel) TableName() string {
	return "movies"
}

// MovieRepository implements movie.Repository interface
// and provides PostgreSQL full-text search.
type MovieRepository struct {
	db *gorm.DB
}

// NewMovieRepository creates a new movie repository
func NewMovieRepository(db *gorm.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

func (r *MovieRepository) Search(ctx context.Context, query string, limit int) ([]movie.Movie, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	const sql = `
SELECT movie_id, title, genres
FROM movies
WHERE search_vector @@ websearch_to_tsquery('english', ?)
ORDER BY ts_rank(search_vector, websearch_to_tsquery('english', ?)) DESC, movie_id
LIMIT ?`

	var models []MovieModel
	if err := r.db.WithContext(ctx).Raw(sql, query, query, limit).Scan(&models).Error; err != nil {
		return nil, err
	}

	movies := make([]movie.Movie, len(models))
	for i, model := range models {
		movies[i] = movie.Movie{
			MovieID: model.MovieID,
			Title:   model.Title,
			Genres:  model.Genres,
		}
	}
	return movies, nil
}
