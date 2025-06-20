package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const oldDBName = "everdell.db"
const newDBName = "games.db"

func main() {
	log.Println("Starting database migration...")

	// 1. Safety Check: Ensure the new database doesn't already exist.
	if _, err := os.Stat(newDBName); err == nil {
		log.Printf("Migration STOPPED: The new database '%s' already exists. It seems the migration has already been run.", newDBName)
		return
	}

	// 2. Check for the old database file.
	if _, err := os.Stat(oldDBName); os.IsNotExist(err) {
		log.Printf("Migration SKIPPED: The old database '%s' was not found. Nothing to migrate.", oldDBName)
		return
	}

	// 3. Rename the old database file to the new name.
	log.Printf("Renaming '%s' to '%s'...", oldDBName, newDBName)
	err := os.Rename(oldDBName, newDBName)
	if err != nil {
		log.Fatalf("FATAL: Could not rename database file: %v", err)
	}
	log.Println("Database file renamed successfully.")

	// 4. Open the newly renamed database.
	db, err := sql.Open("sqlite3", "./"+newDBName)
	if err != nil {
		log.Fatalf("FATAL: Failed to open the new database '%s': %v", newDBName, err)
	}
	defer db.Close()

	// 5. Add the new 'game_type' column to the 'games' table.
	log.Println("Adding 'game_type' column to the 'games' table...")
	alterStatement := `ALTER TABLE games ADD COLUMN game_type TEXT;`
	_, err = db.Exec(alterStatement)
	if err != nil {
		// Note: This might fail if the column somehow already exists.
		// The script is designed to run once, so this is a fatal error.
		log.Fatalf("FATAL: Failed to alter the games table: %v", err)
	}
	log.Println("'game_type' column added successfully.")

	// 6. Update all existing rows to set the game_type to 'everdell'.
	log.Println("Updating existing records to set game_type = 'everdell'...")
	updateStatement := `UPDATE games SET game_type = 'everdell';`
	res, err := db.Exec(updateStatement)
	if err != nil {
		log.Fatalf("FATAL: Failed to update existing game records: %v", err)
	}

	rowsAffected, _ := res.RowsAffected()
	log.Printf("Successfully updated %d records.", rowsAffected)
	fmt.Println("\n-----------------------------------------")
	fmt.Println("✅ Migration completed successfully!")
	fmt.Printf("Your data is now in '%s' and is ready for the new application.\n", newDBName)
	fmt.Println("-----------------------------------------")
}
