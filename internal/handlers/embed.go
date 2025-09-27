package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AlbertGhazaly/Steganography-on-Audio-Files-with-Multiple-LSB-Method/internal/utils"
)

func EmbedHandler(w http.ResponseWriter, r *http.Request) {
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
	useKeyForPosition := r.FormValue("use_key_for_position") == "true"
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

	secretFile, secretHeader, err := r.FormFile("secret_file")
	if err != nil {
		utils.SendError(w, "Secret file is required", http.StatusBadRequest)
		return
	}
	defer secretFile.Close()

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

	// For now, just return the same MP3 file
	// TODO: Implement actual steganography embedding

	mp3Dst.Close()
	mp3Data, err := os.ReadFile(mp3Path)
	if err != nil {
		utils.SendError(w, "Failed to read MP3 file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"stego_%s\"", mp3Header.Filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(mp3Data)))

	w.Write(mp3Data)

	log.Printf("Embed operation: key=%s, encryption=%v, keyPosition=%v, lsb=%d, mp3=%s, secret=%s",
		key, useEncryption, useKeyForPosition, lsbBits, mp3Header.Filename, secretHeader.Filename)
}
