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

func (l *LSBSteganography) EmbedMessage(mp3Data, message []byte, bits int) ([]byte, error) {
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

	l.embedData(result[l.headerSize:], messageWithHeader, bits)

	return result, nil
}

func (l *LSBSteganography) addLengthHeader(message []byte) []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, uint32(len(message)))

	buf.Write(message)

	return buf.Bytes()
}

func (l *LSBSteganography) embedData(carrier, message []byte, bits int) {

	totalMessageBits := len(message) * 8

	for i := 0; i < totalMessageBits; i++ {
		messageByteIndex := i / 8
		messageBitPosition := 7 - (i % 8)

		messageBit := (message[messageByteIndex] >> messageBitPosition) & 1

		carrierByteIndex := (i * bits) / 8
		if carrierByteIndex >= len(carrier) {
			break
		}

		carrierBitOffset := (i * bits) % 8
		carrierBitPosition := 7 - carrierBitOffset

		bitClearMask := byte(^(1 << carrierBitPosition))
		carrier[carrierByteIndex] &= bitClearMask

		if messageBit == 1 {
			carrier[carrierByteIndex] |= (1 << carrierBitPosition)
		}

		for j := 1; j < bits && carrierBitPosition-j >= 0; j++ {
			carrier[carrierByteIndex] &= byte(^(1 << (carrierBitPosition - j)))
		}
	}
}

func (l *LSBSteganography) ExtractMessage(mp3Data []byte, bits int) ([]byte, error) {
	if bits < 1 || bits > 4 {
		return nil, ErrInvalidBitCount
	}

	if len(mp3Data) <= l.headerSize {
		return nil, ErrInvalidMP3Format
	}

	lengthBytes := make([]byte, 4)
	for i := range lengthBytes {
		lengthBytes[i] = 0
	}

	stegoData := mp3Data[l.headerSize:]
	mask := byte((1 << bits) - 1)

	for i := 0; i < 32; i++ {
		byteIndex := i / 8
		bitPosition := 7 - (i % 8)

		carrierByteIndex := (i * bits) / 8
		carrierBitOffset := (i * bits) % 8

		var extractedBits byte
		if carrierBitOffset+bits <= 8 {
			extractedBits = (stegoData[carrierByteIndex] >> (8 - carrierBitOffset - bits)) & mask
		} else {
			bitsFromFirst := 8 - carrierBitOffset
			extractedBits = (stegoData[carrierByteIndex] & ((1 << bitsFromFirst) - 1)) << (bits - bitsFromFirst)

			if carrierByteIndex+1 < len(stegoData) {
				bitsFromSecond := bits - bitsFromFirst
				secondByteBits := (stegoData[carrierByteIndex+1] >> (8 - bitsFromSecond)) & ((1 << bitsFromSecond) - 1)
				extractedBits |= secondByteBits
			}
		}

		bit := (extractedBits >> (bits - 1)) & 1

		if bit == 1 {
			lengthBytes[byteIndex] |= (1 << bitPosition)
		}
	}

	messageLength := binary.BigEndian.Uint32(lengthBytes)

	capacity := CalculateCapacity(len(mp3Data), bits)
	if messageLength == 0 || int(messageLength) > capacity-4 {
		return nil, ErrInvalidMP3Format
	}

	message := make([]byte, messageLength)
	for i := range message {
		message[i] = 0
	}

	for i := 0; i < int(messageLength)*8; i++ {
		byteIndex := i / 8
		bitPosition := 7 - (i % 8)

		carrierByteIndex := ((i + 32) * bits) / 8
		carrierBitOffset := ((i + 32) * bits) % 8

		if carrierByteIndex >= len(stegoData) {
			break
		}

		var extractedBits byte
		if carrierBitOffset+bits <= 8 {
			extractedBits = (stegoData[carrierByteIndex] >> (8 - carrierBitOffset - bits)) & mask
		} else {
			bitsFromFirst := 8 - carrierBitOffset
			extractedBits = (stegoData[carrierByteIndex] & ((1 << bitsFromFirst) - 1)) << (bits - bitsFromFirst)

			if carrierByteIndex+1 < len(stegoData) {
				bitsFromSecond := bits - bitsFromFirst
				secondByteBits := (stegoData[carrierByteIndex+1] >> (8 - bitsFromSecond)) & ((1 << bitsFromSecond) - 1)
				extractedBits |= secondByteBits
			}
		}

		bit := (extractedBits >> (bits - 1)) & 1

		if bit == 1 {
			message[byteIndex] |= (1 << bitPosition)
		}
	}

	return message, nil
}

func (l *LSBSteganography) extractData(carrier, output []byte, bits int) {
	mask := byte((1 << bits) - 1)

	for i := range output {
		output[i] = 0
	}

	totalOutBits := len(output) * 8
	currentOutBit := 0

	for i := 0; i < len(carrier) && currentOutBit < totalOutBits; i++ {
		extractedBits := carrier[i] & mask

		outByteIndex := currentOutBit / 8
		outBitOffset := currentOutBit % 8

		if outBitOffset+bits <= 8 {
			output[outByteIndex] |= extractedBits << (8 - outBitOffset - bits)
		} else {
			firstPartBits := 8 - outBitOffset
			secondPartBits := bits - firstPartBits

			output[outByteIndex] |= (extractedBits >> secondPartBits) & ((1 << firstPartBits) - 1)

			if outByteIndex+1 < len(output) {
				output[outByteIndex+1] |= (extractedBits & ((1 << secondPartBits) - 1)) << (8 - secondPartBits)
			}
		}

		currentOutBit += bits
	}
}
