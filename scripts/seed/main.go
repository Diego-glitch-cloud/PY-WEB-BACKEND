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

// Variables globales del script

var db *sql.DB
var apiKey string



func main() {
	// Cargar .env desde la raiz del proyecto
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no se encontro .env, usando variables de entorno del sistema")
	}

	apiKey = os.Getenv("RAWG_API_KEY")
	if apiKey == "" {
		log.Fatal("RAWG_API_KEY no esta configurada")
	}

	// Conectar a PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("error al abrir la base de datos: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("error al conectar con la base de datos: %v", err)
	}
	fmt.Println("conectado a PostgreSQL")

	// Obtener juegos de RAWG con sus detalles
	fmt.Println("obteniendo juegos de RAWG...")
	games, err := fetchGamesWithDetails(25)
	if err != nil {
		log.Fatalf("error al obtener juegos de RAWG: %v", err)
	}
	fmt.Printf("obtenidos %d juegos\n\n", len(games))

	// Insertar cada juego en la base de datos
	inserted, skipped := 0, 0
	for i, g := range games {
		fmt.Printf("[%d/%d] %s... ", i+1, len(games), g.Name)

		ok, err := seedGame(g)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		if ok {
			fmt.Println("insertado")
			inserted++
		} else {
			fmt.Println("omitido (ya existe)")
			skipped++
		}
	}

	fmt.Printf("\nresultado final: %d insertados, %d omitidos\n", inserted, skipped)
}

// Funciones de llamada a RAWG 

// fetchGamesWithDetails obtiene la lista de juegos y luego el detalle de cada uno.
// El endpoint de lista no incluye description ni developers
func fetchGamesWithDetails(count int) ([]rawgDetail, error) {
	url := fmt.Sprintf(
		"%s/games?key=%s&tags=singleplayer&page_size=%d&ordering=-rating&exclude_additions=true",
		rawgBase, apiKey, count,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error al llamar RAWG list: %w", err)
	}
	defer resp.Body.Close()

	var list rawgListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("error al decodificar lista: %w", err)
	}

	var details []rawgDetail
	for _, item := range list.Results {
		time.Sleep(250 * time.Millisecond) // respetar rate limit de RAWG

		detail, err := fetchDetail(item.ID)
		if err != nil {
			log.Printf("aviso: no se pudo obtener detalle de '%s', usando datos basicos\n", item.Name)
			details = append(details, rawgDetail{rawgListItem: item})
			continue
		}
		details = append(details, *detail)
	}

	return details, nil
}

// fetchDetail obtiene el detalle completo de un juego por su ID de RAWG.
func fetchDetail(id int) (*rawgDetail, error) {
	url := fmt.Sprintf("%s/games/%d?key=%s", rawgBase, id, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var detail rawgDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, err
	}
	return &detail, nil
}