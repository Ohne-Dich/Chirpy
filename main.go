package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Ohne-Dich/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error - no connection to the database: %s", err)
	}
	const filepathRoot = "."
	const port = "8081"

	var apiCfg apiConfig
	apiCfg.dbQueries = database.New(db)

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("/screenshots/", http.StripPrefix("/screenshots/", http.FileServer(http.Dir("./screenshots"))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricsResett)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)
	mux.HandleFunc("GET /api/download_zip", handlerDownloadZip)
	mux.HandleFunc("POST /api/upload_screenshot", apiCfg.handlerUploadScreenshot)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}
