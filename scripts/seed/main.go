package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const rawgBase = "https://api.rawg.io/api"

// ---- Tipos para deserializar la respuesta de RAWG ----

type rawgListResponse struct {
	Results []rawgListItem `json:"results"`
}

type rawgListItem struct {
	ID              int                   `json:"id"`
	Name            string                `json:"name"`
	Released        string                `json:"released"`
	BackgroundImage string                `json:"background_image"`
	Genres          []rawgGenre           `json:"genres"`
	Platforms       []rawgPlatformWrapper `json:"platforms"`
}

// rawgDetail extiende rawgListItem con campos que solo devuelve el endpoint de detalle
type rawgDetail struct {
	rawgListItem
	DescriptionRaw string          `json:"description_raw"`
	Developers     []rawgDeveloper `json:"developers"`
}

type rawgGenre struct {
	Name string `json:"name"`
}

type rawgPlatformWrapper struct {
	Platform struct {
		Name string `json:"name"`
	} `json:"platform"`
}

type rawgDeveloper struct {
	Name string `json:"name"`
}