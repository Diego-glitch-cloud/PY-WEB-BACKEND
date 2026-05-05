package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Diego-glitch-cloud/PY-WEB-BACKEND/db"
	"github.com/Diego-glitch-cloud/PY-WEB-BACKEND/models"
)

// CreateRating godoc
// POST /games/:id/rating
func CreateRating(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "el ID debe ser un numero entero",
			Field:   "id",
		})
		return
	}

	var exists bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM games WHERE id = $1)", gameID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: fmt.Sprintf("no existe un juego con ID %d", gameID),
		})
		return
	}

	var input models.RatingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: "el score debe ser un numero entre 1 y 5",
			Field:   "score",
		})
		return
	}

	var r models.Rating
	err = db.DB.QueryRow(`
		INSERT INTO ratings (game_id, score)
		VALUES ($1, $2)
		RETURNING id, game_id, score, created_at`,
		gameID, input.Score,
	).Scan(&r.ID, &r.GameID, &r.Score, &r.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "db_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, r)
}

// GetRatings godoc
// GET /games/:id/rating
func GetRatings(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "el ID debe ser un numero entero",
			Field:   "id",
		})
		return
	}

	var exists bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM games WHERE id = $1)", gameID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: fmt.Sprintf("no existe un juego con ID %d", gameID),
		})
		return
	}

	// Obtener promedio y conteo
	var response models.RatingResponse
	response.GameID = gameID

	err = db.DB.QueryRow(`
		SELECT COALESCE(AVG(score), 0), COUNT(*)
		FROM ratings WHERE game_id = $1`, gameID,
	).Scan(&response.Average, &response.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "db_error",
			Message: err.Error(),
		})
		return
	}

	// Obtener lista de ratings individuales
	rows, err := db.DB.Query(`
		SELECT id, game_id, score, created_at
		FROM ratings WHERE game_id = $1
		ORDER BY created_at DESC`, gameID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "db_error",
			Message: err.Error(),
		})
		return
	}
	defer rows.Close()

	response.Scores = []models.Rating{}
	for rows.Next() {
		var r models.Rating
		if err := rows.Scan(&r.ID, &r.GameID, &r.Score, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "scan_error",
				Message: err.Error(),
			})
			return
		}
		response.Scores = append(response.Scores, r)
	}

	c.JSON(http.StatusOK, response)
}