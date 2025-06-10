package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type CropParams struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (cfg *apiConfig) handlerUploadScreenshot(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20) // 50MB

	cropJson := r.FormValue("cropRect")
	crop := CropParams{X: 0, Y: 0, Width: 0, Height: 0}
	if err := json.Unmarshal([]byte(cropJson), &crop); err != nil {
		http.Error(w, "Crop-Daten ungültig", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["screenshots"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()
		if len(files) == 0 {
			http.Error(w, "Keine Dateien erhalten", http.StatusBadRequest)
			return
		}

		sessionPath := "uploads/session"
		os.MkdirAll(sessionPath, os.ModePerm)

		for _, fh := range files {
			file, err := fh.Open()
			if err != nil {
				continue
			}
			defer file.Close()

			img, _, err := image.Decode(file)
			if err != nil {
				continue
			}

			bounds := img.Bounds()
			if crop.X < 0 || crop.Y < 0 || crop.X+crop.Width > bounds.Dx() || crop.Y+crop.Height > bounds.Dy() {
				continue
			}

			cropRect := image.Rect(crop.X, crop.Y, crop.X+crop.Width, crop.Y+crop.Height)
			cropped := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(cropRect)

			outPath := filepath.Join(sessionPath, "cropped_"+fh.Filename)
			outFile, err := os.Create(outPath)
			if err != nil {
				continue
			}
			defer outFile.Close()

			png.Encode(outFile, cropped)
		}
	}
	w.Write([]byte(fmt.Sprintf("%d Screenshot(s) verarbeitet und gespeichert", len(files))))

}

func handlerDownloadZip(w http.ResponseWriter, r *http.Request) {
	sessionPath := "uploads/session"
	files, err := os.ReadDir(sessionPath)
	if err != nil || len(files) == 0 {
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

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=downloads.zip")
	w.Write(zipBuffer.Bytes())

	// Nach dem Download löschen
	go func() {
		time.Sleep(5 * time.Second)
		for _, file := range files {
			os.Remove(filepath.Join(sessionPath, file.Name()))
		}
	}()
}
