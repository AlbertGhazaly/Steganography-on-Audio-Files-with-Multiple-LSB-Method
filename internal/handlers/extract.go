package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	// For now, just return a dummy text file
	// TODO: Implement actual steganography extraction
	dummyContent := fmt.Sprintf("Extracted secret file\nKey: %s\nEncryption: %v\nKey Position: %v\nLSB Bits: %d\nFrom: %s",
		key, useEncryption, useKeyForPosition, lsbBits, mp3Header.Filename)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"extracted_secret.txt\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(dummyContent)))

	w.Write([]byte(dummyContent))

	log.Printf("Extract operation: key=%s, encryption=%v, keyPosition=%v, lsb=%d, mp3=%s",
		key, useEncryption, useKeyForPosition, lsbBits, mp3Header.Filename)
}
