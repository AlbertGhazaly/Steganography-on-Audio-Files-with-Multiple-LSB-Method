package stego

type LSBResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	EmbeddedFile []byte `json:"embedded_file"`
}

func NewSuccessResponse(embeddedFile []byte) *LSBResponse {
	return &LSBResponse{
		Success:      true,
		Message:      "Successfully embedded message into MP3 file",
		EmbeddedFile: embeddedFile,
	}
}

func NewErrorResponse(message string) *LSBResponse {
	return &LSBResponse{
		Success:      false,
		Message:      message,
		EmbeddedFile: nil,
	}
}

func CalculateCapacity(mp3FileSize, bits int) int {
	dataSize := mp3FileSize - 1024
	if dataSize <= 0 {
		return 0
	}

	capacityInBits := dataSize * bits
	return capacityInBits / 8
}
