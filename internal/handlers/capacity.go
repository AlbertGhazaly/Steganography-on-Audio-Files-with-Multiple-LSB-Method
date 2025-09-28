package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/stego"
	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/utils"
)

type CapacityResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	CapacityBytes    int    `json:"capacity_bytes"`
	CapacityReadable string `json:"capacity_readable"`
	FrameCount       int    `json:"frame_count"`
	Method           string `json:"method"`
}

func CapacityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(100 << 20) // 100 MB limit
	if err != nil {
		utils.SendError(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	mp3File, mp3Header, err := r.FormFile("mp3_file")
	if err != nil {
		utils.SendError(w, "MP3 file is required", http.StatusBadRequest)
		return
	}
	defer mp3File.Close()

	tempDir := "./temp"
	os.MkdirAll(tempDir, 0755)

	mp3Path := filepath.Join(tempDir, mp3Header.Filename)
	mp3Dst, err := os.Create(mp3Path)
	if err != nil {
		utils.SendError(w, "Failed to save MP3 file", http.StatusInternalServerError)
		return
	}
	defer mp3Dst.Close()
	defer os.Remove(mp3Path)

	_, err = io.Copy(mp3Dst, mp3File)
	if err != nil {
		utils.SendError(w, "Failed to save MP3 file", http.StatusInternalServerError)
		return
	}
	mp3Dst.Close()

	mp3Data, err := os.ReadFile(mp3Path)
	if err != nil {
		utils.SendError(w, "Failed to read MP3 file", http.StatusInternalServerError)
		return
	}

	headerStego := stego.NewHeaderSteganography()
	capacity, frameCount, err := headerStego.CalculateCapacity(mp3Data)
	if err != nil {
		utils.SendError(w, "Failed to calculate capacity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	capacityReadable := formatBytes(capacity)

	response := CapacityResponse{
		Success:          true,
		Message:          "Capacity calculated successfully",
		CapacityBytes:    capacity,
		CapacityReadable: capacityReadable,
		FrameCount:       frameCount,
		Method:           "MP3 Header Steganography",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d bytes", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
