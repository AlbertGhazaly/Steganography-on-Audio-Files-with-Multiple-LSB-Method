package stego

import "errors"

var (
	ErrEmptyMessageFile     = errors.New("message file cannot be empty")
	ErrEmptyMP3File         = errors.New("MP3 file cannot be empty")
	ErrInvalidBitCount      = errors.New("bits must be between 1 and 4")
	ErrInsufficientCapacity = errors.New("MP3 file is too small to embed the message")
	ErrInvalidMP3Format     = errors.New("invalid MP3 file format")
)

type LSBRequest struct {
	MessageFile []byte `json:"message_file"`
	Mp3File     []byte `json:"mp3_file"`
	Bits        int    `json:"bits"`
}

func (req *LSBRequest) Validate() error {
	if len(req.MessageFile) == 0 {
		return ErrEmptyMessageFile
	}

	if len(req.Mp3File) == 0 {
		return ErrEmptyMP3File
	}

	if req.Bits < 1 || req.Bits > 4 {
		return ErrInvalidBitCount
	}

	return nil
}
