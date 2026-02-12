package main

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"hexagon/pkg/config"
	"hexagon/postgres"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const defaultMovieLensURL = "https://files.grouplens.org/datasets/movielens/ml-latest-small.zip"

func main() {
	var (
		csvPath string
		zipURL  string
		limit   int
	)

	flag.StringVar(&csvPath, "csv", "", "Path to movies.csv (skip download)")
	flag.StringVar(&zipURL, "url", defaultMovieLensURL, "MovieLens zip URL")
	flag.IntVar(&limit, "limit", 0, "Limit number of rows to import (0 = all)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	db, err := postgres.NewConnection(postgres.Options{
		DBName:   cfg.DB.Name,
		DBUser:   cfg.DB.User,
		Password: cfg.DB.Pass,
		Host:     cfg.DB.Host,
		Port:     fmt.Sprintf("%d", cfg.DB.Port),
		SSLMode:  cfg.DB.EnableSSL,
	})
	if err != nil {
		slog.Error("cannot open postgres connection", "error", err)
		os.Exit(1)
	}

	cleanup := func() {}
	if csvPath == "" {
		path, c, err := downloadAndExtract(zipURL)
		if err != nil {
			slog.Error("failed to download dataset", "error", err)
			os.Exit(1)
		}
		csvPath = path
		cleanup = c
	}
	defer cleanup()

	count, err := importMovies(context.Background(), db, csvPath, limit)
	if err != nil {
		slog.Error("import failed", "error", err)
		os.Exit(1)
	}

	slog.Info("import completed", "rows", count)
}

func downloadAndExtract(zipURL string) (string, func(), error) {
	if zipURL == "" {
		return "", func() {}, errors.New("dataset url is empty")
	}

	tmpDir, err := os.MkdirTemp("", "movielens-")
	if err != nil {
		return "", func() {}, err
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	zipPath := filepath.Join(tmpDir, "dataset.zip")
	if err := downloadFile(zipURL, zipPath); err != nil {
		cleanup()
		return "", func() {}, err
	}

	csvPath, err := extractMoviesCSV(zipPath, tmpDir)
	if err != nil {
		cleanup()
		return "", func() {}, err
	}

	return csvPath, cleanup, nil
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url) // nolint: noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractMoviesCSV(zipPath, destDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, file := range r.File {
		if !strings.HasSuffix(file.Name, "movies.csv") {
			continue
		}

		src, err := file.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		destPath := filepath.Join(destDir, filepath.Base(file.Name))
		out, err := os.Create(destPath)
		if err != nil {
			return "", err
		}

		if _, err := io.Copy(out, src); err != nil {
			_ = out.Close()
			return "", err
		}
		if err := out.Close(); err != nil {
			return "", err
		}

		return destPath, nil
	}

	return "", errors.New("movies.csv not found in zip")
}

func importMovies(ctx context.Context, db *gorm.DB, csvPath string, limit int) (int, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	idxMovieID, idxTitle, idxGenres, err := parseMovieCSVHeader(reader)
	if err != nil {
		return 0, err
	}

	stmt := `
INSERT INTO movies (movie_id, title, genres)
VALUES (?, ?, ?)
ON CONFLICT (movie_id) DO UPDATE SET
	title = EXCLUDED.title,
	genres = EXCLUDED.genres
`

	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	count := 0
	for limit <= 0 || count < limit {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_ = tx.Rollback()
			return count, err
		}
		movieID, title, genres, ok := parseMovieRecord(record, idxMovieID, idxTitle, idxGenres)
		if !ok {
			continue
		}

		if err := tx.Exec(stmt, movieID, title, genres).Error; err != nil {
			_ = tx.Rollback()
			return count, err
		}

		count++
	}

	if err := tx.Commit().Error; err != nil {
		return count, err
	}

	return count, nil
}

func parseMovieCSVHeader(reader *csv.Reader) (int, int, int, error) {
	header, err := reader.Read()
	if err != nil {
		return 0, 0, 0, err
	}

	idxMovieID, idxTitle, idxGenres := -1, -1, -1
	for i, name := range header {
		switch strings.TrimSpace(name) {
		case "movieId":
			idxMovieID = i
		case "title":
			idxTitle = i
		case "genres":
			idxGenres = i
		}
	}
	if idxMovieID == -1 || idxTitle == -1 || idxGenres == -1 {
		return 0, 0, 0, errors.New("missing required columns in csv header")
	}

	return idxMovieID, idxTitle, idxGenres, nil
}

func parseMovieRecord(record []string, idxMovieID, idxTitle, idxGenres int) (int, string, string, bool) {
	if idxMovieID >= len(record) || idxTitle >= len(record) || idxGenres >= len(record) {
		return 0, "", "", false
	}

	movieID, err := strconv.Atoi(strings.TrimSpace(record[idxMovieID]))
	if err != nil {
		return 0, "", "", false
	}
	title := strings.TrimSpace(record[idxTitle])
	genres := strings.TrimSpace(record[idxGenres])
	return movieID, title, genres, true
}
