package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Diego-glitch-cloud/PY-WEB-BACKEND/db"
	"github.com/Diego-glitch-cloud/PY-WEB-BACKEND/models"
)

// allowedSorts mapea ?sort= a columnas SQL reales — previene SQL injection.
var allowedSorts = map[string]string{
	"title":        "g.title",
	"release_year": "g.release_year",
	"genre":        "gen.name",
	"created_at":   "g.created_at",
	"updated_at":   "g.updated_at",
}

// baseSelect es el SELECT con JOINs reutilizado en GetGames y GetGame.
const baseSelect = `
	SELECT
		g.id, g.title,
		COALESCE(gen.name, '') AS genre,
		COALESCE(plt.name, '') AS platform,
		COALESCE(dev.name, '') AS developer,
		g.release_year,
		g.description,
		g.image_url,
		AVG(r.score) AS avg_rating,
		g.created_at,
		g.updated_at
	FROM games g
	LEFT JOIN genres gen ON g.genre_id = gen.id
	LEFT JOIN platforms plt ON g.platform_id = plt.id
	LEFT JOIN developers dev ON g.developer_id = dev.id
	LEFT JOIN ratings r ON g.id = r.game_id`

// scanner abstrae *sql.Row y *sql.Rows para reutilizar scanGame en ambos casos.
type scanner interface {
	Scan(dest ...any) error
}

// scanGame escanea una fila de baseSelect en un GameResponse.
func scanGame(s scanner) (models.GameResponse, error) {
	var g models.GameResponse
	var releaseYear sql.NullInt64
	var avgRating sql.NullFloat64

	err := s.Scan(
		&g.ID, &g.Title, &g.Genre, &g.Platform, &g.Developer,
		&releaseYear, &g.Description, &g.ImageURL,
		&avgRating, &g.CreatedAt, &g.UpdatedAt,
	)
	if releaseYear.Valid {
		y := int(releaseYear.Int64)
		g.ReleaseYear = &y
	}
	if avgRating.Valid {
		g.AvgRating = &avgRating.Float64
	}
	return g, err
}

// getOrCreateLookup inserta un valor en una tabla de lookup y devuelve su ID.
// Los nombres de tabla y columna son siempre literales del código, nunca input del usuario.
func getOrCreateLookup(table, column, value string) (*int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var id int
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES ($1)
		 ON CONFLICT (%s) DO UPDATE SET %s = EXCLUDED.%s
		 RETURNING id`,
		table, column, column, column, column,
	)
	if err := db.DB.QueryRow(query, value).Scan(&id); err != nil {
		return nil, err
	}
	return &id, nil
}

// GetGames godoc
// GET /games?q=&sort=&order=&page=&limit=
func GetGames(c *gin.Context) {
	q := c.Query("q")
	sortParam := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}
	sortCol, ok := allowedSorts[sortParam]
	if !ok {
		sortCol = "g.created_at"
	}

	// Construir WHERE y args dinamicamente
	args := []any{}
	where := ""
	if q != "" {
		args = append(args, "%"+q+"%")
		where = fmt.Sprintf("WHERE g.title ILIKE $%d", len(args))
	}

	// Contar total para la paginacion
	countQuery := fmt.Sprintf(
		"SELECT COUNT(*) FROM games g LEFT JOIN genres gen ON g.genre_id = gen.id %s",
		where,
	)
	var total int
	if err := db.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "db_error",
			Message: err.Error(),
		})
		return
	}

	// Agregar LIMIT y OFFSET a los args
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		%s
		%s
		GROUP BY g.id, gen.name, plt.name, dev.name
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		baseSelect, where, sortCol, order, len(args)-1, len(args),
	)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "db_error",
			Message: err.Error(),
		})
		return
	}
	defer rows.Close()

	games := []models.GameResponse{}
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "scan_error",
				Message: err.Error(),
			})
			return
		}
		games = append(games, g)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       games,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetGame godoc
// GET /games/:id
func GetGame(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "el ID debe ser un numero entero",
			Field:   "id",
		})
		return
	}

	query := fmt.Sprintf(`
		%s
		WHERE g.id = $1
		GROUP BY g.id, gen.name, plt.name, dev.name`,
		baseSelect,
	)

	g, err := scanGame(db.DB.QueryRow(query, id))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: fmt.Sprintf("no existe un juego con ID %d", id),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "db_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, g)
}

func CreateGame(c *gin.Context)  {}
func UpdateGame(c *gin.Context)  {}
func DeleteGame(c *gin.Context)  {}
func UploadImage(c *gin.Context) {}
