package main

import (
	"archive/zip"
	"bytes"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func (cfg *apiConfig) handlerUploadScreenshot(w http.ResponseWriter, r *http.Request) {
	// Session-Ordner leeren
	files, _ := os.ReadDir("uploads/session")
	for _, f := range files {
		os.Remove(filepath.Join("uploads/session", f.Name()))
	}
	// Max. Upload-Größe auf 10MB beschränken
	r.ParseMultipartForm(10 << 20)
	cropStr := r.FormValue("crop")
	cropHeight := 40 // default fallback

	if cropStr != "" {
		if val, err := strconv.Atoi(cropStr); err == nil && val >= 0 {
			cropHeight = val
		}
	}

	file, handler, err := r.FormFile("screenshot")
	if err != nil {
		http.Error(w, "Fehler beim Lesen des Uploads", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Bild dekodieren
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Ungültiges Bildformat", http.StatusUnsupportedMediaType)
		return
	}

	// Zuschneiden (z. B. Taskleiste unten entfernen)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	cropRect := image.Rect(0, 0, width, height-cropHeight)
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	// Sicherstellen, dass Zielordner existiert
	sessionPath := "uploads/session"
	os.MkdirAll(sessionPath, os.ModePerm)

	// Neue Datei speichern
	outputPath := filepath.Join(sessionPath, "cropped_"+handler.Filename)
	outFile, err := os.Create(outputPath)
	if err != nil {
		http.Error(w, "Fehler beim Speichern", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	err = png.Encode(outFile, croppedImg)
	if err != nil {
		http.Error(w, "Fehler beim Encoden", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Screenshot gespeichert unter: " + outputPath + "\n"))
}

func handlerDownloadZip(w http.ResponseWriter, r *http.Request) {
	sessionPath := "uploads/session"

	files, err := os.ReadDir(sessionPath)
	if err != nil {
		http.Error(w, "Fehler beim Lesen des Session-Ordners", http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		http.Error(w, "Keine Bilder zum Herunterladen gefunden", http.StatusNotFound)
		return
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fullPath := filepath.Join(sessionPath, file.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		f, err := zipWriter.Create(file.Name())
		if err != nil {
			continue
		}
		f.Write(content)
	}

	zipWriter.Close()

	// HTTP-Header setzen
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"session_bilder.zip\"")
	w.Write(zipBuffer.Bytes())

	// Nach dem Download: löschen
	for _, file := range files {
		os.Remove(filepath.Join(sessionPath, file.Name()))
	}
}
