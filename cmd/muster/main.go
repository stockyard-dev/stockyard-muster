package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/stockyard-dev/stockyard-muster/internal/license"
	"github.com/stockyard-dev/stockyard-muster/internal/server"
	"github.com/stockyard-dev/stockyard-muster/internal/store"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Printf("muster %s\n", version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && (os.Args[1] == "--health" || os.Args[1] == "health") {
		fmt.Println("ok")
		os.Exit(0)
	}
	log.SetFlags(log.Ltime | log.Lshortfile)
	port := 8910
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil { port = n }
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" { dataDir = "./data" }
	licenseKey := os.Getenv("MUSTER_LICENSE_KEY")
	licInfo, licErr := license.Validate(licenseKey, "muster")
	if licenseKey != "" && licErr != nil {
		log.Printf("[license] WARNING: %v — running in free tier", licErr)
		licInfo = nil
	}
	limits := server.LimitsFor(licInfo)
	if licInfo != nil && licInfo.IsPro() {
		log.Printf("  License:   Pro (%s)", licInfo.CustomerID)
	} else {
		log.Printf("  License:   Free tier (set MUSTER_LICENSE_KEY to unlock Pro)")
	}
	db, err := store.Open(dataDir)
	if err != nil { log.Fatalf("database: %v", err) }
	defer db.Close()
	log.Printf("")
	log.Printf("  Stockyard Muster %s", version)
	log.Printf("  On-call:        http://localhost:%d/api/oncall", port)
	log.Printf("  API:            http://localhost:%d/api/incidents", port)
	log.Printf("  Dashboard:      http://localhost:%d/ui", port)
	log.Printf("")
	srv := server.New(db, port, limits)
	if err := srv.Start(); err != nil { log.Fatalf("server: %v", err) }
}
