package models

import "time"

// GameInput es lo que recibe el API al crear o editar un juego.
// El cliente envía genre, platform y developer como strings.
type GameInput struct {
	Title       string `json:"title"        binding:"required,max=255"`
	Genre       string `json:"genre"        binding:"omitempty,max=100"`
	Platform    string `json:"platform"     binding:"omitempty,max=100"`
	Developer   string `json:"developer"    binding:"omitempty,max=255"`
	ReleaseYear *int   `json:"release_year" binding:"omitempty,min=1970,max=2030"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

// GameResponse es lo que devuelve el API.
// genre, platform y developer vienen como strings via JOIN con las tablas de lookup.
type GameResponse struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Genre       string   `json:"genre"`
	Platform    string   `json:"platform"`
	Developer   string   `json:"developer"`
	ReleaseYear *int     `json:"release_year"`
	Description string   `json:"description"`
	ImageURL    string   `json:"image_url"`
	AvgRating   *float64 `json:"avg_rating"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaginatedResponse envuelve una lista de juegos con metadata de paginacion.
type PaginatedResponse struct {
	Data       []GameResponse `json:"data"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int            `json:"total"`
	TotalPages int            `json:"total_pages"`
}

// Rating representa una calificacion individual de un videojuego.
type Rating struct {
	ID        int       `json:"id"`
	GameID    int       `json:"game_id"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

// RatingInput es lo que recibe POST /games/:id/rating.
type RatingInput struct {
	Score int `json:"score" binding:"required,min=1,max=5"`
}

// RatingResponse es el resumen de ratings de un juego (promedio + lista).
type RatingResponse struct {
	GameID  int      `json:"game_id"`
	Average float64  `json:"average"`
	Count   int      `json:"count"`
	Scores  []Rating `json:"scores"`
}

// ErrorResponse es el formato estandar de error del API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}