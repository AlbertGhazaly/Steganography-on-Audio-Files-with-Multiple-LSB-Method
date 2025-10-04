package stego

import (
	"bytes"
	"encoding/binary"
)

type LSBSteganography struct {
	headerSize int
}

func NewLSBSteganography() *LSBSteganography {
	return &LSBSteganography{
		headerSize: 1024,
	}
}

func (l *LSBSteganography) calculateKeyOffset(key string) int {
	if key == "" {
		return 0
	}

	sum := 0
	for i := 0; i < len(key); i += 2 {
		sum += int(key[i])
	}
	return sum
}

func (l *LSBSteganography) EmbedMessage(mp3Data, message []byte, bits int) ([]byte, error) {
	return l.EmbedMessageWithKey(mp3Data, message, bits, "", false)
}

func (l *LSBSteganography) EmbedMessageWithKey(mp3Data, message []byte, bits int, key string, useKeyForPosition bool) ([]byte, error) {
	return l.EmbedMessageWithMetadata(mp3Data, message, bits, key, useKeyForPosition, false, "", "")
}

func (l *LSBSteganography) EmbedMessageWithMetadata(mp3Data, message []byte, bits int, key string, useKeyForPosition bool, useEncryption bool, originalFilename string, fileType string) ([]byte, error) {
	if bits < 1 || bits > 4 {
		return nil, ErrInvalidBitCount
	}

	if len(mp3Data) <= l.headerSize {
		return nil, ErrInvalidMP3Format
	}

	metadata := &EmbedMetadata{
		UseEncryption:     useEncryption,
		UseKeyForPosition: useKeyForPosition,
		LSBBits:           bits,
		OriginalFilename:  originalFilename,
		FileType:          fileType,
		SecretMessageSize: len(message),
	}

	metadataBytes, err := SerializeMetadata(metadata, key)
	if err != nil {
		return nil, err
	}

	var payload bytes.Buffer

	binary.Write(&payload, binary.BigEndian, uint32(len(metadataBytes)))
	payload.Write(metadataBytes)

	binary.Write(&payload, binary.BigEndian, uint32(len(message)))
	payload.Write(message)

	payloadData := payload.Bytes()

	capacity := CalculateCapacity(len(mp3Data), bits)
	if len(payloadData) > capacity {
		return nil, ErrInsufficientCapacity
	}

	result := make([]byte, len(mp3Data))
	copy(result, mp3Data)

	var offset int
	if useKeyForPosition && key != "" {
		offset = l.calculateKeyOffset(key)
	}

	l.embedDataWithOffset(result[l.headerSize:], payloadData, bits, offset)

	return result, nil
}

func (l *LSBSteganography) embedDataWithOffset(carrier, message []byte, bits int, offset int) {
	mask := byte((1 << bits) - 1)
	totalMessageBits := len(message) * 8

	startOffset := offset % len(carrier)
	availableCarrier := len(carrier) - startOffset
	carrierBitCapacity := availableCarrier * bits

	if totalMessageBits > carrierBitCapacity {
		totalMessageBits = carrierBitCapacity
	}

	for i := 0; i < totalMessageBits; i += bits {
		carrierIndex := startOffset + (i / bits)
		if carrierIndex >= len(carrier) {
			break
		}

		messageBitStart := i
		var chunk byte = 0
		for b := 0; b < bits && messageBitStart+b < totalMessageBits; b++ {
			msgByte := message[(messageBitStart+b)/8]
			msgBitPos := 7 - ((messageBitStart + b) % 8)
			msgBit := (msgByte >> msgBitPos) & 1
			chunk = (chunk << 1) | msgBit
		}

		carrier[carrierIndex] &= ^mask

		carrier[carrierIndex] |= chunk & mask
	}
}

func (l *LSBSteganography) ExtractMessage(mp3Data []byte, bits int) ([]byte, error) {
	return l.ExtractMessageWithKey(mp3Data, bits, "", false)
}

func (l *LSBSteganography) ExtractMessageWithKey(mp3Data []byte, bits int, key string, useKeyForPosition bool) ([]byte, error) {
	if bits < 1 || bits > 4 {
		return nil, ErrInvalidBitCount
	}
	if len(mp3Data) <= l.headerSize {
		return nil, ErrInvalidMP3Format
	}

	stegoData := mp3Data[l.headerSize:]
	mask := byte((1 << bits) - 1)

	var offset int
	if useKeyForPosition && key != "" {
		offset = l.calculateKeyOffset(key)
	}
	startOffset := offset % len(stegoData)

	lengthBits := 32
	lengthBytes := make([]byte, 4)

	for i := 0; i < lengthBits; i += bits {
		carrierIndex := startOffset + (i / bits)
		if carrierIndex >= len(stegoData) {
			break
		}

		bitChunk := stegoData[carrierIndex] & mask

		for b := 0; b < bits && i+b < lengthBits; b++ {
			bitValue := (bitChunk >> (bits - 1 - b)) & 1
			byteIndex := (i + b) / 8
			bitPos := 7 - ((i + b) % 8)

			if bitValue == 1 {
				lengthBytes[byteIndex] |= (1 << bitPos)
			}
		}
	}

	messageLength := binary.BigEndian.Uint32(lengthBytes)
	totalMessageBits := int(messageLength) * 8

	message := make([]byte, messageLength)

	for i := 0; i < totalMessageBits; i += bits {
		carrierIndex := startOffset + ((i + lengthBits) / bits)
		if carrierIndex >= len(stegoData) {
			break
		}

		bitChunk := stegoData[carrierIndex] & mask

		for b := 0; b < bits && i+b < totalMessageBits; b++ {
			bitValue := (bitChunk >> (bits - 1 - b)) & 1
			byteIndex := (i + b) / 8
			bitPos := 7 - ((i + b) % 8)

			if bitValue == 1 {
				message[byteIndex] |= (1 << bitPos)
			}
		}
	}

	return message, nil
}

type ExtractResult struct {
	Message          []byte
	Metadata         *EmbedMetadata
	OriginalFilename string
	FileType         string
}

func (l *LSBSteganography) ExtractMessageWithMetadata(mp3Data []byte, key string) (*ExtractResult, error) {
	if len(mp3Data) <= l.headerSize {
		return nil, ErrInvalidMP3Format
	}

	stegoData := mp3Data[l.headerSize:]

	for bits := 1; bits <= 4; bits++ {
		mask := byte((1 << bits) - 1)

		for _, useKeyForPos := range []bool{false, true} {
			var offset int
			if useKeyForPos && key != "" {
				offset = l.calculateKeyOffset(key)
			}
			startOffset := offset % len(stegoData)

			metadataLengthBits := 32
			metadataLengthBytes := make([]byte, 4)

			success := true
			for i := 0; i < metadataLengthBits && success; i += bits {
				carrierIndex := startOffset + (i / bits)
				if carrierIndex >= len(stegoData) {
					success = false
					break
				}

				bitChunk := stegoData[carrierIndex] & mask

				for b := 0; b < bits && i+b < metadataLengthBits; b++ {
					bitValue := (bitChunk >> (bits - 1 - b)) & 1
					byteIndex := (i + b) / 8
					bitPos := 7 - ((i + b) % 8)

					if bitValue == 1 {
						metadataLengthBytes[byteIndex] |= (1 << bitPos)
					}
				}
			}

			if !success {
				continue
			}

			metadataLength := binary.BigEndian.Uint32(metadataLengthBytes)

			if metadataLength == 0 || metadataLength > 10000 {
				continue
			}

			metadataBytes := make([]byte, metadataLength)
			totalMetadataBits := int(metadataLength) * 8
			metadataStartBit := 32

			for i := 0; i < totalMetadataBits; i += bits {
				carrierIndex := startOffset + ((i + metadataStartBit) / bits)
				if carrierIndex >= len(stegoData) {
					success = false
					break
				}

				bitChunk := stegoData[carrierIndex] & mask

				for b := 0; b < bits && i+b < totalMetadataBits; b++ {
					bitValue := (bitChunk >> (bits - 1 - b)) & 1
					byteIndex := (i + b) / 8
					bitPos := 7 - ((i + b) % 8)

					if bitValue == 1 {
						metadataBytes[byteIndex] |= (1 << bitPos)
					}
				}
			}

			if !success {
				continue
			}

			metadata, _, err := DeserializeMetadata(metadataBytes, key)
			if err != nil {
				continue
			}

			if metadata.LSBBits != bits || metadata.UseKeyForPosition != useKeyForPos {
				continue
			}

			messageStartBit := metadataStartBit + totalMetadataBits

			messageLengthBits := 32
			messageLengthBytes := make([]byte, 4)

			for i := 0; i < messageLengthBits; i += bits {
				carrierIndex := startOffset + ((i + messageStartBit) / bits)
				if carrierIndex >= len(stegoData) {
					success = false
					break
				}

				bitChunk := stegoData[carrierIndex] & mask

				for b := 0; b < bits && i+b < messageLengthBits; b++ {
					bitValue := (bitChunk >> (bits - 1 - b)) & 1
					byteIndex := (i + b) / 8
					bitPos := 7 - ((i + b) % 8)

					if bitValue == 1 {
						messageLengthBytes[byteIndex] |= (1 << bitPos)
					}
				}
			}

			if !success {
				continue
			}

			messageLength := binary.BigEndian.Uint32(messageLengthBytes)

			if int(messageLength) != metadata.SecretMessageSize {
				continue
			}

			message := make([]byte, messageLength)
			totalMessageBits := int(messageLength) * 8
			messageDataStartBit := messageStartBit + messageLengthBits

			for i := 0; i < totalMessageBits; i += bits {
				carrierIndex := startOffset + ((i + messageDataStartBit) / bits)
				if carrierIndex >= len(stegoData) {
					success = false
					break
				}

				bitChunk := stegoData[carrierIndex] & mask

				for b := 0; b < bits && i+b < totalMessageBits; b++ {
					bitValue := (bitChunk >> (bits - 1 - b)) & 1
					byteIndex := (i + b) / 8
					bitPos := 7 - ((i + b) % 8)

					if bitValue == 1 {
						message[byteIndex] |= (1 << bitPos)
					}
				}
			}

			if !success {
				continue
			}

			return &ExtractResult{
				Message:          message,
				Metadata:         metadata,
				OriginalFilename: metadata.OriginalFilename,
				FileType:         metadata.FileType,
			}, nil
		}
	}

	return nil, ErrNoSteganographicData
}
