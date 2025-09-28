package stego

type HeaderResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	EmbeddedFile []byte `json:"embedded_file"`
}

func NewSuccessResponse(embeddedFile []byte) *HeaderResponse {
	return &HeaderResponse{
		Success:      true,
		Message:      "Successfully embedded message into MP3 file using header steganography",
		EmbeddedFile: embeddedFile,
	}
}

func NewErrorResponse(message string) *HeaderResponse {
	return &HeaderResponse{
		Success:      false,
		Message:      message,
		EmbeddedFile: nil,
	}
}

type LSBResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	EmbeddedFile []byte `json:"embedded_file"`
}

func CalculateCapacity(mp3FileSize, bits int) int {
	dataSize := mp3FileSize - 1024
	if dataSize <= 0 {
		return 0
	}

	capacityInBits := dataSize * bits
	return capacityInBits / 8
}

func PaperCalculateCapacity(mp3FileSize, bits int) int {
	return CalculateCapacity(mp3FileSize, bits) - 100
}
