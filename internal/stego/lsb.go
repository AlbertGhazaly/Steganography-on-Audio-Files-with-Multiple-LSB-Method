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
	if bits < 1 || bits > 4 {
		return nil, ErrInvalidBitCount
	}

	if len(mp3Data) <= l.headerSize {
		return nil, ErrInvalidMP3Format
	}

	capacity := CalculateCapacity(len(mp3Data), bits)

	if len(message)+4 > capacity {
		return nil, ErrInsufficientCapacity
	}

	result := make([]byte, len(mp3Data))
	copy(result, mp3Data)

	messageWithHeader := l.addLengthHeader(message)

	var offset int
	if useKeyForPosition && key != "" {
		offset = l.calculateKeyOffset(key)
	}

	l.embedDataWithOffset(result[l.headerSize:], messageWithHeader, bits, offset)

	return result, nil
}

func (l *LSBSteganography) addLengthHeader(message []byte) []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, uint32(len(message)))

	buf.Write(message)

	return buf.Bytes()
}

func (l *LSBSteganography) embedData(carrier, message []byte, bits int) {
	l.embedDataWithOffset(carrier, message, bits, 0)
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
