package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/crypto"
	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/stego"
	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/utils"
)

func ExtractHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(100 << 20) // 100 MB limit
	if err != nil {
		utils.SendError(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	key := r.FormValue("key")
	useEncryption := r.FormValue("use_encryption") == "true"
	// useKeyForPosition := r.FormValue("use_key_for_position") == "true"
	mode := r.FormValue("mode")
	if mode == "" {
		mode = "paper"
	}
	lsbBitsStr := r.FormValue("lsb_bits")

	lsbBits, err := strconv.Atoi(lsbBitsStr)
	if err != nil || lsbBits < 1 || lsbBits > 4 {
		utils.SendError(w, "LSB bits must be between 1 and 4", http.StatusBadRequest)
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

	// Use header-based steganography instead of LSB
	headerStego := stego.NewHeaderSteganography()
	extractedData, err := headerStego.ExtractMessage(mp3Data)
	if err != nil {
		utils.SendError(w, "Failed to extract secret data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if useEncryption && key != "" {
		extractedData = crypto.VigenereDecrypt(extractedData, key)
	}

	contentType := http.DetectContentType(extractedData)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\"extracted_secret\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(extractedData)))

	w.Write(extractedData)

	log.Printf("Extract operation: method=header, mp3=%s", mp3Header.Filename)
}
