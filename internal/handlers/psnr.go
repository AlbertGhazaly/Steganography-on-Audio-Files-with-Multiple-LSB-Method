package handlers

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/utils"
)

// PSNRResponse represents the response structure for PSNR calculation
type PSNRResponse struct {
	PSNR         float64 `json:"psnr"`
	MSE          float64 `json:"mse"`
	MaxSignal    float64 `json:"max_signal"`
	OriginalSize int     `json:"original_size"`
	ModifiedSize int     `json:"modified_size"`
}

// PSNRHandler calculates the PSNR between original and modified audio files
func PSNRHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(100 << 20) // 100 MB limit
	if err != nil {
		utils.SendError(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get original MP3 file
	originalFile, originalHeader, err := r.FormFile("original_file")
	if err != nil {
		utils.SendError(w, "Original MP3 file is required", http.StatusBadRequest)
		return
	}
	defer originalFile.Close()

	// Get modified MP3 file
	modifiedFile, modifiedHeader, err := r.FormFile("modified_file")
	if err != nil {
		utils.SendError(w, "Modified MP3 file is required", http.StatusBadRequest)
		return
	}
	defer modifiedFile.Close()

	// Create temporary directory
	tempDir := "./temp"
	os.MkdirAll(tempDir, 0755)

	// Save the original file temporarily
	originalPath := filepath.Join(tempDir, "original_"+originalHeader.Filename)
	originalDst, err := os.Create(originalPath)
	if err != nil {
		utils.SendError(w, "Failed to save original file", http.StatusInternalServerError)
		return
	}
	defer originalDst.Close()
	defer os.Remove(originalPath)

	_, err = io.Copy(originalDst, originalFile)
	if err != nil {
		utils.SendError(w, "Failed to save original file", http.StatusInternalServerError)
		return
	}
	originalDst.Close()

	// Save the modified file temporarily
	modifiedPath := filepath.Join(tempDir, "modified_"+modifiedHeader.Filename)
	modifiedDst, err := os.Create(modifiedPath)
	if err != nil {
		utils.SendError(w, "Failed to save modified file", http.StatusInternalServerError)
		return
	}
	defer modifiedDst.Close()
	defer os.Remove(modifiedPath)

	_, err = io.Copy(modifiedDst, modifiedFile)
	if err != nil {
		utils.SendError(w, "Failed to save modified file", http.StatusInternalServerError)
		return
	}
	modifiedDst.Close()

	// Read the original file
	originalData, err := os.ReadFile(originalPath)
	if err != nil {
		utils.SendError(w, "Failed to read original file", http.StatusInternalServerError)
		return
	}

	// Read the modified file
	modifiedData, err := os.ReadFile(modifiedPath)
	if err != nil {
		utils.SendError(w, "Failed to read modified file", http.StatusInternalServerError)
		return
	}

	// Calculate PSNR
	psnr, mse, maxVal := calculatePSNR(originalData, modifiedData)

	// Create response
	response := PSNRResponse{
		PSNR:         psnr,
		MSE:          mse,
		MaxSignal:    maxVal,
		OriginalSize: len(originalData),
		ModifiedSize: len(modifiedData),
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("PSNR calculation: PSNR=%.2f dB, MSE=%.6f, original=%s, modified=%s",
		psnr, mse, originalHeader.Filename, modifiedHeader.Filename)
}

// calculatePSNR calculates the PSNR between two byte arrays
func calculatePSNR(original, modified []byte) (float64, float64, float64) {
	// Get minimum length of both arrays
	minLen := len(original)
	if len(modified) < minLen {
		minLen = len(modified)
	}

	// Calculate MSE
	var sumSquaredError float64
	var maxVal byte
	for i := 0; i < minLen; i++ {
		// Find maximum sample value (needed for PSNR)
		if original[i] > maxVal {
			maxVal = original[i]
		}

		// Calculate squared error
		diff := float64(original[i]) - float64(modified[i])
		sumSquaredError += diff * diff
	}

	// Calculate mean squared error
	mse := sumSquaredError / float64(minLen)

	// Convert max value to float64
	maxValF64 := float64(maxVal)
	if maxValF64 == 0 {
		maxValF64 = 255.0 // Default to 8-bit max if no signal found
	}

	// Calculate PSNR
	psnr := 20*math.Log10(maxValF64) - 10*math.Log10(mse)
	if mse == 0 {
		psnr = float64(100.0) // If MSE is 0, files are identical, set a high PSNR value
	}

	return psnr, mse, float64(maxVal)
}
