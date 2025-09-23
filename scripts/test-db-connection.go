package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Database connection parameters
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "tvl_aggregator")
	sslmode := getEnv("DB_SSL_MODE", "disable")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	fmt.Println("Testing database connection...")
	fmt.Printf("Host: %s:%s\n", host, port)
	fmt.Printf("Database: %s\n", dbname)
	fmt.Printf("User: %s\n", user)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connection successful!")

	// Test basic queries
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query database version: %v", err)
	}

	fmt.Printf("📊 PostgreSQL Version: %s\n", version)

	// Check if our tables exist
	var tableCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name IN ('chains', 'protocols', 'tvl_snapshots')
	`).Scan(&tableCount)
	if err != nil {
		log.Fatalf("Failed to check tables: %v", err)
	}

	fmt.Printf("📋 Found %d main tables in database\n", tableCount)

	if tableCount > 0 {
		// Count chains
		var chainCount int
		err = db.QueryRow("SELECT COUNT(*) FROM chains").Scan(&chainCount)
		if err == nil {
			fmt.Printf("🔗 Chains configured: %d\n", chainCount)
		}

		// Count protocols
		var protocolCount int
		err = db.QueryRow("SELECT COUNT(*) FROM protocols").Scan(&protocolCount)
		if err == nil {
			fmt.Printf("🏛️  Protocols configured: %d\n", protocolCount)
		}
	} else {
		fmt.Println("⚠️  No tables found. Run migrations to set up the schema.")
	}

	fmt.Println("\n✅ Database test completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}