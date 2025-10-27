package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Config contiene la configuración de la base de datos
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetConfigFromEnv obtiene la configuración desde variables de entorno
func GetConfigFromEnv() *Config {
	return &Config{
		Host:     getEnvOrDefault("POSTGRES_HOST", getEnvOrDefault("DB_HOST", "localhost")),
		Port:     getEnvOrDefault("POSTGRES_PORT", getEnvOrDefault("DB_PORT", "5432")),
		User:     getEnvOrDefault("POSTGRES_USER", getEnvOrDefault("DB_USER", "postgres")),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", getEnvOrDefault("DB_PASSWORD", "postgres")),
		DBName:   getEnvOrDefault("POSTGRES_DB", getEnvOrDefault("DB_NAME", "banca_en_linea")),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}
}

// GetDSN construye el Data Source Name para PostgreSQL
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Connect establece una conexión a la base de datos
func Connect(config *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Verificar la conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Configurar el pool de conexiones
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

// RunMigrations ejecuta las migraciones de la base de datos
func RunMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	// Intentar ejecutar migraciones
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// Si hay un error de dirty database, intentar forzar la versión
		if err.Error() == "Dirty database version 1. Fix and force version." {
			log.Println("Database is in dirty state, forcing version...")
			if forceErr := m.Force(1); forceErr != nil {
				return fmt.Errorf("could not force database version: %w", forceErr)
			}
			// Intentar nuevamente después de forzar
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("could not run migrations after force: %w", err)
			}
		} else {
			return fmt.Errorf("could not run migrations: %w", err)
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}

// getEnvOrDefault obtiene una variable de entorno o devuelve un valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}