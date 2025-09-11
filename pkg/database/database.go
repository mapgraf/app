package database

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/nalgeon/redka"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	dbMap = make(map[string]*redka.DB)
	dbMu  sync.RWMutex // Mutex to handle concurrent access
)

func init() {

	// Handle termination signals to close all database connections
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan // Block until a termination signal is received

		// Close all database connections before exiting
		dbMu.Lock()
		defer dbMu.Unlock()
		for filename, db := range dbMap {
			if err := db.Close(); err != nil {
				log.DefaultLogger.Error("Error closing database connection for", "filename", filename, "error", err)
			}
		}
	}()
}

// Retrieves the database connection for a given filename, opening it if necessary.
func GetDB(filename string) (*redka.DB, error) {
	dbMu.RLock()
	db, exists := dbMap[filename]
	dbMu.RUnlock()

	if exists {
		return db, nil // Return existing connection
	}

	// Open a new connection if it doesn't exist
	dbMu.Lock()
	defer dbMu.Unlock()

	// Check again in case another goroutine opened it
	if db, exists := dbMap[filename]; exists {
		return db, nil
	}

	opts := redka.Options{
		DriverName: "sqlite",
	}

	var err error
	db, err = redka.Open(filename, &opts)
	if err != nil {
		return nil, err
	}

	dbMap[filename] = db
	return db, nil
}
