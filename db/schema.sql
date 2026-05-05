-- ============================================================
-- Video Game Tracker — Database Schema
-- Ejecutar una sola vez para inicializar la base de datos.
-- Compatible con PostgreSQL 14+
-- ============================================================

-- Trigger function para auto-actualizar updated_at en cada UPDATE
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ------------------------------------------------------------
-- Tablas de lookup (normalizacion 3FN — evitan strings repetidos)
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS genres (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS platforms (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS developers (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

-- ------------------------------------------------------------
-- Tabla principal de videojuegos
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS games (
    id           SERIAL PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    genre_id     INTEGER REFERENCES genres(id)    ON DELETE SET NULL,
    platform_id  INTEGER REFERENCES platforms(id) ON DELETE SET NULL,
    developer_id INTEGER REFERENCES developers(id) ON DELETE SET NULL,
    release_year INTEGER CHECK (release_year >= 1970 AND release_year <= 2030),
    description  TEXT,
    -- Almacena URLs: RAWG CDN para datos seeded, /uploads/<file> para imagenes subidas
    image_url    TEXT,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP TRIGGER IF EXISTS games_updated_at ON games;
CREATE TRIGGER games_updated_at
    BEFORE UPDATE ON games
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- ------------------------------------------------------------
-- Tabla de ratings (relacion N:1 con games)
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS ratings (
    id         SERIAL PRIMARY KEY,
    game_id    INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    score      INTEGER NOT NULL CHECK (score >= 1 AND score <= 5),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ------------------------------------------------------------
-- Indices para los patrones de consulta mas comunes
-- ------------------------------------------------------------

CREATE INDEX IF NOT EXISTS idx_games_title        ON games(title);
CREATE INDEX IF NOT EXISTS idx_games_genre_id     ON games(genre_id);
CREATE INDEX IF NOT EXISTS idx_games_release_year ON games(release_year);
CREATE INDEX IF NOT EXISTS idx_ratings_game_id    ON ratings(game_id);
