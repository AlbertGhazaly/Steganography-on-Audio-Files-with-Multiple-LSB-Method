package stego

import "errors"

var (
	ErrEmptyMessageFile     = errors.New("message file cannot be empty")
	ErrEmptyMP3File         = errors.New("MP3 file cannot be empty")
	ErrInvalidBitCount      = errors.New("bits must be between 1 and 4")
	ErrInsufficientCapacity = errors.New("MP3 file is too small to embed the message")
	ErrInvalidMP3Format     = errors.New("invalid MP3 file format")
	ErrNoValidFrames        = errors.New("no valid MP3 frames found")
	ErrEmbedDataTooLarge    = errors.New("secret data too large for MP3 header capacity")
	ErrInvalidMetadata      = errors.New("invalid metadata format")
	ErrWrongKey             = errors.New("incorrect key provided - unable to decrypt encrypted metadata")
	ErrNoSteganographicData = errors.New("no steganographic data found in this MP3 file")
)

type HeaderRequest struct {
	MessageFile []byte `json:"message_file"`
	Mp3File     []byte `json:"mp3_file"`
	Filename    string `json:"filename"`
}

func (req *HeaderRequest) Validate() error {
	if len(req.MessageFile) == 0 {
		return ErrEmptyMessageFile
	}

	if len(req.Mp3File) == 0 {
		return ErrEmptyMP3File
	}

	return nil
}

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
