package handlers

import (
	"fmt"
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

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		utils.SendError(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	key := r.FormValue("key")
	method := r.FormValue("method")
	if method == "" {
		method = "lsb"
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

	var extractedData []byte
	var originalFilename string
	var fileType string
	var metadata *stego.EmbedMetadata

	if method == "header" {
		headerStego := stego.NewHeaderSteganography()
		extractedData, err = headerStego.ExtractMessage(mp3Data)
		if err != nil {
			utils.SendError(w, "Failed to extract secret data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		originalFilename = "extracted_secret"
		fileType = http.DetectContentType(extractedData)
	} else {
		lsbStego := stego.NewLSBSteganography()
		result, err := lsbStego.ExtractMessageWithMetadata(mp3Data, key)
		if err != nil {
			var errorMsg string
			var statusCode int

			switch err {
			case stego.ErrWrongKey:
				errorMsg = "Incorrect key provided. Please check your key and try again. If the file was embedded with encryption, you must provide the correct key used during embedding."
				statusCode = http.StatusBadRequest
			case stego.ErrNoSteganographicData:
				errorMsg = "No steganographic data found in this MP3 file. Please make sure you uploaded the correct file that contains embedded data."
				statusCode = http.StatusBadRequest
			case stego.ErrInvalidMetadata:
				errorMsg = "Invalid or corrupted steganographic data found. The file may be damaged or not properly embedded."
				statusCode = http.StatusBadRequest
			case stego.ErrInvalidMP3Format:
				errorMsg = "Invalid MP3 file format. Please upload a valid MP3 file."
				statusCode = http.StatusBadRequest
			default:
				errorMsg = "Failed to extract secret data: " + err.Error()
				statusCode = http.StatusInternalServerError
			}

			utils.SendError(w, errorMsg, statusCode)
			return
		}

		extractedData = result.Message
		metadata = result.Metadata
		originalFilename = result.OriginalFilename
		fileType = result.FileType

		if metadata.UseEncryption && key != "" {
			log.Printf("Applying decryption based on metadata")
			extractedData = crypto.VigenereDecrypt(extractedData, key)
		}

		log.Printf("Extracted with metadata: filename=%s, type=%s, size=%d, encryption=%t, keyPos=%t, lsbBits=%d",
			originalFilename, fileType, metadata.SecretMessageSize, metadata.UseEncryption,
			metadata.UseKeyForPosition, metadata.LSBBits)
	}

	if originalFilename == "" {
		originalFilename = "extracted_secret"
	}

	contentType := fileType
	if contentType == "" {
		contentType = http.DetectContentType(extractedData)
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", originalFilename))
	w.Header().Set("Content-Length", strconv.Itoa(len(extractedData)))

	if metadata != nil {
		w.Header().Set("X-Original-Filename", originalFilename)
		w.Header().Set("X-File-Type", fileType)
		w.Header().Set("X-Secret-Size", strconv.Itoa(metadata.SecretMessageSize))
		w.Header().Set("X-Used-Encryption", strconv.FormatBool(metadata.UseEncryption))
		w.Header().Set("X-Used-Key-Position", strconv.FormatBool(metadata.UseKeyForPosition))
		w.Header().Set("X-LSB-Bits", strconv.Itoa(metadata.LSBBits))
	}

	w.Write(extractedData)

	log.Printf("Extract operation: method=%s, mp3=%s, extracted=%s", method, mp3Header.Filename, originalFilename)
}
