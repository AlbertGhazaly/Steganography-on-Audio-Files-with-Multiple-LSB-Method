package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type MP3FrameHeader struct {
	Sync       uint16
	Version    uint8
	Layer      uint8
	Protection uint8
	Bitrate    uint8
	SampleRate uint8
	Padding    uint8
	Private    uint8
	Channel    uint8
	ModeExt    uint8
	Copyright  uint8
	Original   uint8
	Emphasis   uint8
	Size       int
}

var bitrateTable = []int{
	0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0,
}

var sampleRateTable = []int{
	44100, 48000, 32000, 0,
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("MP3 Header Steganography Tool")
		fmt.Println("Usage:")
		fmt.Println("  Embed: go run mp3_header_stego.go embed <input.mp3> <secret_file> [output.mp3]")
		fmt.Println("  Extract: go run mp3_header_stego.go extract <embedded.mp3> <output_secret_file>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run mp3_header_stego.go embed ../cmd/test.mp3 secret.txt embedded.mp3")
		fmt.Println("  go run mp3_header_stego.go extract embedded.mp3 extracted_secret.txt")
		os.Exit(1)
	}

	mode := os.Args[1]
	mp3File := os.Args[2]
	secretFile := os.Args[3]

	switch mode {
	case "embed":
		outputFile := "embedded_output.mp3"
		if len(os.Args) > 4 {
			outputFile = os.Args[4]
		}
		err := embedIntoHeaders(mp3File, secretFile, outputFile)
		if err != nil {
			fmt.Printf("Error embedding: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Successfully embedded %s into %s -> %s\n", secretFile, mp3File, outputFile)

	case "extract":
		err := extractFromHeaders(mp3File, secretFile)
		if err != nil {
			fmt.Printf("Error extracting: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Successfully extracted secret from %s -> %s\n", mp3File, secretFile)

	default:
		fmt.Printf("Invalid mode: %s. Use 'embed' or 'extract'\n", mode)
		os.Exit(1)
	}
}

func embedIntoHeaders(mp3Path, secretPath, outputPath string) error {
	mp3Data, err := readFile(mp3Path)
	if err != nil {
		return fmt.Errorf("failed to read MP3 file: %v", err)
	}

	secretData, err := readFile(secretPath)
	if err != nil {
		return fmt.Errorf("failed to read secret file: %v", err)
	}

	fmt.Printf("MP3 file: %s (%d bytes)\n", mp3Path, len(mp3Data))
	fmt.Printf("Secret file: %s (%d bytes)\n", secretPath, len(secretData))

	dataStart := skipID3Tag(mp3Data)
	if dataStart > 0 {
		fmt.Printf("Skipped ID3 tag: %d bytes\n", dataStart)
	}

	frames, offsets, err := findMP3Frames(mp3Data[dataStart:])
	if err != nil {
		return fmt.Errorf("failed to parse MP3 frames: %v", err)
	}

	if len(frames) == 0 {
		return fmt.Errorf("no valid MP3 frames found")
	}

	for i := range offsets {
		offsets[i] += dataStart
	}

	fmt.Printf("Found %d MP3 frames\n", len(frames))

	capacity := calculateHeaderCapacity(frames)
	requiredSize := len(secretData) + 8
	fmt.Printf("Header capacity: %d bytes\n", capacity)
	fmt.Printf("Required space: %d bytes\n", requiredSize)

	if requiredSize > capacity {
		return fmt.Errorf("secret file too large: need %d bytes, have %d bytes capacity",
			requiredSize, capacity)
	}

	result, err := embedDataInHeaders(mp3Data, secretData, frames, offsets, filepath.Base(secretPath))
	if err != nil {
		return fmt.Errorf("failed to embed data: %v", err)
	}

	err = writeFile(outputPath, result)
	if err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

func extractFromHeaders(mp3Path, outputPath string) error {
	mp3Data, err := readFile(mp3Path)
	if err != nil {
		return fmt.Errorf("failed to read MP3 file: %v", err)
	}

	fmt.Printf("MP3 file: %s (%d bytes)\n", mp3Path, len(mp3Data))

	dataStart := skipID3Tag(mp3Data)
	if dataStart > 0 {
		fmt.Printf("Skipped ID3 tag: %d bytes\n", dataStart)
	}

	frames, offsets, err := findMP3Frames(mp3Data[dataStart:])
	if err != nil {
		return fmt.Errorf("failed to parse MP3 frames: %v", err)
	}

	if len(frames) == 0 {
		return fmt.Errorf("no valid MP3 frames found")
	}

	for i := range offsets {
		offsets[i] += dataStart
	}

	fmt.Printf("Found %d MP3 frames\n", len(frames))

	secretData, originalFilename, err := extractDataFromHeaders(mp3Data, frames, offsets)
	if err != nil {
		return fmt.Errorf("failed to extract data: %v", err)
	}

	fmt.Printf("Extracted %d bytes of secret data\n", len(secretData))
	if originalFilename != "" {
		fmt.Printf("Original filename: %s\n", originalFilename)
	}

	err = writeFile(outputPath, secretData)
	if err != nil {
		return fmt.Errorf("failed to write extracted file: %v", err)
	}

	return nil
}

func embedDataInHeaders(mp3Data []byte, secretData []byte, frames []*MP3FrameHeader, offsets []int, filename string) ([]byte, error) {
	result := make([]byte, len(mp3Data))
	copy(result, mp3Data)

	filenameBytes := []byte(filename)
	if len(filenameBytes) > 255 {
		filenameBytes = filenameBytes[:255]
	}

	payload := make([]byte, 0)
	dataLen := len(secretData)
	payload = append(payload, byte(dataLen>>24), byte(dataLen>>16), byte(dataLen>>8), byte(dataLen))
	filenameLen := len(filenameBytes)
	payload = append(payload, byte(filenameLen>>24), byte(filenameLen>>16), byte(filenameLen>>8), byte(filenameLen))
	payload = append(payload, filenameBytes...)
	payload = append(payload, secretData...)

	fmt.Printf("Payload size: %d bytes (including metadata)\n", len(payload))

	bitIndex := 0
	payloadIndex := 0

	for frameIdx := range frames {
		if payloadIndex >= len(payload) {
			break
		}

		frameOffset := offsets[frameIdx]

		safeBitPositions := []struct {
			offset int
			mask   byte
			shift  int
		}{
			{2, 0x01, 0},
			{3, 0x08, 3},
			{3, 0x04, 2},
		}

		for _, pos := range safeBitPositions {
			if payloadIndex >= len(payload) {
				break
			}

			payloadByte := payload[payloadIndex]
			bitToEmbed := (payloadByte >> (7 - bitIndex)) & 1

			byteOffset := frameOffset + pos.offset
			result[byteOffset] = (result[byteOffset] & ^pos.mask) | (bitToEmbed << pos.shift)

			bitIndex++
			if bitIndex == 8 {
				bitIndex = 0
				payloadIndex++
			}
		}
	}

	if payloadIndex < len(payload) {
		return nil, fmt.Errorf("could not embed all data: embedded %d/%d bytes",
			payloadIndex, len(payload))
	}

	fmt.Printf("Successfully embedded %d bytes into %d frame headers\n", len(payload), len(frames))
	return result, nil
}

func extractDataFromHeaders(mp3Data []byte, frames []*MP3FrameHeader, offsets []int) ([]byte, string, error) {
	bitIndex := 0
	currentByte := byte(0)

	state := "metadata"
	metadataBytes := 0
	dataLength := 0
	filenameLength := 0

	var filenameBytes []byte
	var dataBytes []byte

	safeBitPositions := []struct {
		offset int
		mask   byte
		shift  int
	}{
		{2, 0x01, 0},
		{3, 0x08, 3},
		{3, 0x04, 2},
	}

	for frameIdx := range frames {
		frameOffset := offsets[frameIdx]

		for _, pos := range safeBitPositions {
			byteOffset := frameOffset + pos.offset
			extractedBit := (mp3Data[byteOffset] & pos.mask) >> pos.shift

			currentByte = (currentByte << 1) | extractedBit
			bitIndex++

			if bitIndex == 8 {
				switch state {
				case "metadata":
					if metadataBytes < 4 {
						dataLength = (dataLength << 8) | int(currentByte)
					} else if metadataBytes < 8 {
						filenameLength = (filenameLength << 8) | int(currentByte)
					}
					metadataBytes++

					if metadataBytes == 8 {
						if dataLength <= 0 || dataLength > 10*1024*1024 || filenameLength < 0 || filenameLength > 255 {
							return nil, "", fmt.Errorf("invalid metadata: dataLen=%d, filenameLen=%d", dataLength, filenameLength)
						}
						fmt.Printf("Metadata: data_length=%d, filename_length=%d\n", dataLength, filenameLength)

						if filenameLength > 0 {
							state = "filename"
						} else {
							state = "data"
						}
					}

				case "filename":
					filenameBytes = append(filenameBytes, currentByte)
					if len(filenameBytes) >= filenameLength {
						filename := string(filenameBytes)
						fmt.Printf("Extracted filename: %s\n", filename)
						state = "data"
					}

				case "data":
					dataBytes = append(dataBytes, currentByte)
					if len(dataBytes) >= dataLength {
						return dataBytes, string(filenameBytes), nil
					}
				}

				currentByte = 0
				bitIndex = 0
			}
		}
	}

	return nil, "", fmt.Errorf("could not extract complete data: state=%s, got %d/%d bytes",
		state, len(dataBytes), dataLength)
}

func calculateHeaderCapacity(frames []*MP3FrameHeader) int {
	return (len(frames) * 3) / 8
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func skipID3Tag(data []byte) int {
	if len(data) < 10 {
		return 0
	}

	if data[0] == 'I' && data[1] == 'D' && data[2] == '3' {
		size := int(data[6])<<21 | int(data[7])<<14 | int(data[8])<<7 | int(data[9])
		return 10 + size
	}

	return 0
}

func parseMP3Frame(data []byte, offset int) (*MP3FrameHeader, error) {
	if len(data) < offset+4 {
		return nil, fmt.Errorf("not enough data for frame header")
	}

	if data[offset] != 0xFF || (data[offset+1]&0xE0) != 0xE0 {
		return nil, fmt.Errorf("invalid frame sync")
	}

	header := &MP3FrameHeader{}
	b1, b2, b3, b4 := data[offset], data[offset+1], data[offset+2], data[offset+3]

	header.Sync = uint16(b1)<<3 | uint16(b2>>5)
	header.Version = (b2 >> 3) & 0x03
	header.Layer = (b2 >> 1) & 0x03
	header.Protection = b2 & 0x01
	header.Bitrate = (b3 >> 4) & 0x0F
	header.SampleRate = (b3 >> 2) & 0x03
	header.Padding = (b3 >> 1) & 0x01
	header.Private = b3 & 0x01
	header.Channel = (b4 >> 6) & 0x03
	header.ModeExt = (b4 >> 4) & 0x03
	header.Copyright = (b4 >> 3) & 0x01
	header.Original = (b4 >> 2) & 0x01
	header.Emphasis = b4 & 0x03

	if header.Version == 3 && header.Layer == 1 {
		bitrate := bitrateTable[header.Bitrate] * 1000
		sampleRate := sampleRateTable[header.SampleRate]

		if bitrate == 0 || sampleRate == 0 {
			return nil, fmt.Errorf("invalid bitrate or sample rate")
		}

		header.Size = (144*bitrate)/sampleRate + int(header.Padding)
	} else {
		return nil, fmt.Errorf("unsupported MP3 format")
	}

	return header, nil
}

func findMP3Frames(data []byte) ([]*MP3FrameHeader, []int, error) {
	var frames []*MP3FrameHeader
	var offsets []int

	i := 0
	for i < len(data)-4 {
		if data[i] == 0xFF && (data[i+1]&0xE0) == 0xE0 {
			frame, err := parseMP3Frame(data, i)
			if err == nil && frame.Size > 0 && i+frame.Size <= len(data) {
				frames = append(frames, frame)
				offsets = append(offsets, i)
				i += frame.Size
			} else {
				i++
			}
		} else {
			i++
		}
	}

	return frames, offsets, nil
}
