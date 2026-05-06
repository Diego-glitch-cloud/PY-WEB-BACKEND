package seed

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const rawgBase = "https://api.rawg.io/api"

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

var db *sql.DB
var apiKey string

var gamesToSearch = []string{ // lista de juegos para poblar inicialmente la DB
	"The Legend of Zelda: Breath of the Wild",
	"The Legend of Zelda: Tears of the Kingdom",
	"The Witcher 3: Wild Hunt",
	"Red Dead Redemption 2",
	"Elden Ring",
	"Dark Souls",
	"Bloodborne",
	"Sekiro: Shadows Die Twice",
	"The Elder Scrolls V: Skyrim",
	"The Last of Us",
	"The Last of Us Part II",
	"God of War (2018)",
	"God of War Ragnarök",
	"Horizon Zero Dawn",
	"Horizon Forbidden West",
	"Ghost of Tsushima",
	"Uncharted 2: Among Thieves",
	"Uncharted 4: A Thief’s End",
	"Metal Gear Solid 3: Snake Eater",
	"Metal Gear Solid V: The Phantom Pain",
	"BioShock",
	"BioShock Infinite",
	"Half-Life 2",
	"Portal 2",
	"Resident Evil 4",
	"Resident Evil 2 Remake",
	"Dead Space",
	"Dead Space Remake",
	"Super Mario Galaxy",
	"Super Mario Odyssey",
	"Hollow Knight",
	"Celeste",
	"Hades",
	"Undertale",
	"Disco Elysium",
	"Persona 5 Royal",
	"Final Fantasy VII Remake",
	"Final Fantasy X",
	"Final Fantasy XII",
	"Chrono Trigger",
	"Dragon Quest XI S",
	"Nier: Automata",
	"Nier Replicant",
	"Control",
	"Alan Wake 2",
	"Batman: Arkham City",
	"Batman: Arkham Knight",
	"Prince of Persia: The Sands of Time",
	"Assassin’s Creed II",
	"Fallout: New Vegas",
}

func Run(database *sql.DB) {
	db = database

	apiKey = os.Getenv("RAWG_API_KEY")
	if apiKey == "" {
		log.Println("RAWG_API_KEY no esta configurada. Omitiendo Auto-Seed.")
        return
	}

    // Revisar si ya hay juegos para no repetir el proceso lento de 50 juegos
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM games").Scan(&count); err == nil && count > 0 {
		log.Println("La base de datos ya tiene juegos. Omitiendo Auto-Seed.")
		return
	}

	rand.Seed(time.Now().UnixNano())
	log.Println("Iniciando auto-seed de videojuegos...")

	inserted, skipped := 0, 0
	for i, gameName := range gamesToSearch {
		fmt.Printf("[%d/%d] Buscando: %s... ", i+1, len(gamesToSearch), gameName)

		gameDetail, err := searchAndFetchDetail(gameName)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		if gameDetail == nil {
			fmt.Println("No encontrado")
			continue
		}

		ok, err := seedGame(*gameDetail)
		if err != nil {
			fmt.Printf("ERROR al insertar: %v\n", err)
			continue
		}
		if ok {
			fmt.Println("insertado")
			inserted++
		} else {
			fmt.Println("omitido (ya existe)")
			skipped++
		}

		time.Sleep(250 * time.Millisecond) // Respetar rate limit
	}

	fmt.Printf("\nresultado final: %d insertados, %d omitidos\n", inserted, skipped)
}

func searchAndFetchDetail(name string) (*rawgDetail, error) {
	searchURL := fmt.Sprintf("%s/games?key=%s&search=%s&page_size=1", rawgBase, apiKey, strings.ReplaceAll(name, " ", "+"))

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var list rawgListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	if len(list.Results) == 0 {
		return nil, nil
	}

	return fetchDetail(list.Results[0].ID)
}

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

func getOrCreate(table, column, value string) (*int, error) {
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
	if err := db.QueryRow(query, value).Scan(&id); err != nil {
		return nil, err
	}
	return &id, nil
}

func seedGame(g rawgDetail) (bool, error) {
	var exists bool
	if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM games WHERE title = $1)", g.Name).Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	var genreName, platformName, developerName string
	if len(g.Genres) > 0 {
		genreName = g.Genres[0].Name
	}
	if len(g.Platforms) > 0 {
		platformName = g.Platforms[0].Platform.Name
	}
	if len(g.Developers) > 0 {
		developerName = g.Developers[0].Name
	}

	genreID, _ := getOrCreate("genres", "name", genreName)
	platformID, _ := getOrCreate("platforms", "name", platformName)
	developerID, _ := getOrCreate("developers", "name", developerName)

	var releaseYear *int
	if g.Released != "" {
		parts := strings.Split(g.Released, "-")
		if y, err := strconv.Atoi(parts[0]); err == nil {
			releaseYear = &y
		}
	}

	desc := strings.TrimSpace(g.DescriptionRaw)
	if len(desc) > 5000 {
		desc = desc[:5000]
	}

	var gameID int
	err := db.QueryRow(`
		INSERT INTO games (title, genre_id, platform_id, developer_id, release_year, description, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, g.Name, genreID, platformID, developerID, releaseYear, desc, g.BackgroundImage).Scan(&gameID)

	if err == nil {
		// Insertar ratings aleatorios
		numRatings := rand.Intn(15) + 1 // 1 a 15 ratings
		for i := 0; i < numRatings; i++ {
			// Favorecer scores altos para que se vea mejor (3 a 5)
			score := rand.Intn(3) + 3 
			db.Exec("INSERT INTO ratings (game_id, score) VALUES ($1, $2)", gameID, score)
		}
	}

	return err == nil, err
}
