package database

import (
	"database/sql"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/nalgeon/redka"
	driver "modernc.org/sqlite"
	"os"
	"os/signal"
	"syscall"
)

var DB *redka.DB

func init() {

	sql.Register("sqlite3", &driver.Driver{})

	var err error
	DB, err = redka.Open("./public/seed/mapgl-data.db", nil)
	if err != nil {
		log.DefaultLogger.Info("redka fatal err: ", err)
	}

	// Handle termination signals to close the database connection
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan // Block until a termination signal is received

		// Close the database connection before exiting
		if err := DB.Close(); err != nil {
			log.DefaultLogger.Error("Error closing database connection:", err)
		}
	}()

}
