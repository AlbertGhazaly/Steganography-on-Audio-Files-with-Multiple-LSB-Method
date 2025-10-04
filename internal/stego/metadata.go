package stego

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
)

type EmbedMetadata struct {
	UseEncryption     bool `json:"use_encryption"`
	UseKeyForPosition bool `json:"use_key_for_position"`
	LSBBits           int  `json:"lsb_bits"`

	OriginalFilename  string `json:"original_filename"`
	FileType          string `json:"file_type"`
	SecretMessageSize int    `json:"secret_message_size"`
}

func DetectFileType(data []byte, filename string) string {
	contentType := http.DetectContentType(data)

	if contentType == "application/octet-stream" || contentType == "text/plain" {
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".pdf":
			return "application/pdf"
		case ".txt":
			return "text/plain"
		case ".doc":
			return "application/msword"
		case ".docx":
			return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".mp3":
			return "audio/mpeg"
		case ".wav":
			return "audio/wav"
		case ".mp4":
			return "video/mp4"
		case ".zip":
			return "application/zip"
		default:
			return contentType
		}
	}

	return contentType
}

func SerializeMetadata(metadata *EmbedMetadata, key string) ([]byte, error) {
	unencryptedData := struct {
		UseEncryption     bool `json:"use_encryption"`
		UseKeyForPosition bool `json:"use_key_for_position"`
		LSBBits           int  `json:"lsb_bits"`
	}{
		UseEncryption:     metadata.UseEncryption,
		UseKeyForPosition: metadata.UseKeyForPosition,
		LSBBits:           metadata.LSBBits,
	}

	unencryptedJSON, err := json.Marshal(unencryptedData)
	if err != nil {
		return nil, err
	}

	encryptedData := struct {
		OriginalFilename  string `json:"original_filename"`
		FileType          string `json:"file_type"`
		SecretMessageSize int    `json:"secret_message_size"`
	}{
		OriginalFilename:  metadata.OriginalFilename,
		FileType:          metadata.FileType,
		SecretMessageSize: metadata.SecretMessageSize,
	}

	encryptedJSON, err := json.Marshal(encryptedData)
	if err != nil {
		return nil, err
	}

	if metadata.UseEncryption && key != "" {
		encryptedJSON = VigenereEncrypt(encryptedJSON, key)
	}

	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, uint32(len(unencryptedJSON)))
	buf.Write(unencryptedJSON)

	binary.Write(&buf, binary.BigEndian, uint32(len(encryptedJSON)))
	buf.Write(encryptedJSON)

	return buf.Bytes(), nil
}

func DeserializeMetadata(data []byte, key string) (*EmbedMetadata, int, error) {
	if len(data) < 8 {
		return nil, 0, ErrInvalidMetadata
	}

	buf := bytes.NewReader(data)
	totalBytesRead := 0

	var unencryptedSize uint32
	err := binary.Read(buf, binary.BigEndian, &unencryptedSize)
	if err != nil {
		return nil, 0, err
	}
	totalBytesRead += 4

	if len(data) < int(unencryptedSize)+8 {
		return nil, 0, ErrInvalidMetadata
	}

	unencryptedData := make([]byte, unencryptedSize)
	_, err = buf.Read(unencryptedData)
	if err != nil {
		return nil, 0, err
	}
	totalBytesRead += int(unencryptedSize)

	var unencryptedPart struct {
		UseEncryption     bool `json:"use_encryption"`
		UseKeyForPosition bool `json:"use_key_for_position"`
		LSBBits           int  `json:"lsb_bits"`
	}

	err = json.Unmarshal(unencryptedData, &unencryptedPart)
	if err != nil {
		return nil, 0, err
	}

	var encryptedSize uint32
	err = binary.Read(buf, binary.BigEndian, &encryptedSize)
	if err != nil {
		return nil, 0, err
	}
	totalBytesRead += 4

	if len(data) < int(encryptedSize)+totalBytesRead {
		return nil, 0, ErrInvalidMetadata
	}

	encryptedData := make([]byte, encryptedSize)
	_, err = buf.Read(encryptedData)
	if err != nil {
		return nil, 0, err
	}
	totalBytesRead += int(encryptedSize)

	if unencryptedPart.UseEncryption && key != "" {
		encryptedData = VigenereDecrypt(encryptedData, key)
	} else if unencryptedPart.UseEncryption && key == "" {
		return nil, 0, ErrWrongKey
	}

	var encryptedPart struct {
		OriginalFilename  string `json:"original_filename"`
		FileType          string `json:"file_type"`
		SecretMessageSize int    `json:"secret_message_size"`
	}

	err = json.Unmarshal(encryptedData, &encryptedPart)
	if err != nil {
		if unencryptedPart.UseEncryption {
			return nil, 0, ErrWrongKey
		}
		return nil, 0, ErrInvalidMetadata
	}

	if encryptedPart.SecretMessageSize < 0 || encryptedPart.SecretMessageSize > 100*1024*1024 {
		if unencryptedPart.UseEncryption {
			return nil, 0, ErrWrongKey
		}
		return nil, 0, ErrInvalidMetadata
	}

	metadata := &EmbedMetadata{
		UseEncryption:     unencryptedPart.UseEncryption,
		UseKeyForPosition: unencryptedPart.UseKeyForPosition,
		LSBBits:           unencryptedPart.LSBBits,
		OriginalFilename:  encryptedPart.OriginalFilename,
		FileType:          encryptedPart.FileType,
		SecretMessageSize: encryptedPart.SecretMessageSize,
	}

	return metadata, totalBytesRead, nil
}

func VigenereEncrypt(data []byte, key string) []byte {
	if len(key) == 0 {
		return data
	}

	result := make([]byte, len(data))
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	for i := 0; i < len(data); i++ {
		keyByte := keyBytes[i%keyLen]
		result[i] = byte((uint(data[i]) + uint(keyByte)) % 256)
	}

	return result
}

func VigenereDecrypt(data []byte, key string) []byte {
	if len(key) == 0 {
		return data
	}

	result := make([]byte, len(data))
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	for i := 0; i < len(data); i++ {
		keyByte := keyBytes[i%keyLen]
		result[i] = byte((uint(data[i]) + 256 - uint(keyByte)%256) % 256)
	}

	return result
}
