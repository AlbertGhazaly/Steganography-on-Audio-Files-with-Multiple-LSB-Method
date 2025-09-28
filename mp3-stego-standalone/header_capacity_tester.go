package main

import (
	"fmt"
	"io"
	"os"
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

type HeaderEmbeddingInfo struct {
	TotalFrames        int
	UsableHeaderBits   int
	SafeBitsPerFrame   int
	TotalCapacityBits  int
	TotalCapacityBytes int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("MP3 Header Capacity Tester")
		fmt.Println("Usage: go run header_capacity_tester.go <mp3_file>")
		fmt.Println("Example: go run header_capacity_tester.go ../cmd/test.mp3")
		os.Exit(1)
	}

	mp3File := os.Args[1]

	err := analyzeMP3HeaderCapacity(mp3File)
	if err != nil {
		fmt.Printf("Error analyzing MP3: %v\n", err)
		os.Exit(1)
	}
}

func analyzeMP3HeaderCapacity(mp3Path string) error {
	mp3Data, err := readFile(mp3Path)
	if err != nil {
		return fmt.Errorf("failed to read MP3 file: %v", err)
	}

	fmt.Printf("=== MP3 Header Capacity Analysis ===\n")
	fmt.Printf("File: %s\n", mp3Path)
	fmt.Printf("File size: %d bytes (%.2f MB)\n", len(mp3Data), float64(len(mp3Data))/1024/1024)

	dataStart := skipID3Tag(mp3Data)
	if dataStart > 0 {
		fmt.Printf("ID3 tag size: %d bytes (skipped)\n", dataStart)
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

	fmt.Printf("Total MP3 frames found: %d\n", len(frames))

	conservativeInfo := analyzeHeaderEmbeddingCapacity(frames, offsets, mp3Data)

	aggressiveInfo := analyzeAggressiveCapacity(frames)

	displayCapacityResults(conservativeInfo, aggressiveInfo)

	showFrameBreakdown(frames[:min(5, len(frames))], offsets[:min(5, len(offsets))], mp3Data)

	return nil
}

func analyzeHeaderEmbeddingCapacity(frames []*MP3FrameHeader, offsets []int, mp3Data []byte) *HeaderEmbeddingInfo {
	info := &HeaderEmbeddingInfo{
		TotalFrames: len(frames),
	}

	conservativeBits := 3
	info.SafeBitsPerFrame = conservativeBits
	info.TotalCapacityBits = info.TotalFrames * conservativeBits
	info.TotalCapacityBytes = info.TotalCapacityBits / 8

	return info
}

func analyzeAggressiveCapacity(frames []*MP3FrameHeader) *HeaderEmbeddingInfo {
	info := &HeaderEmbeddingInfo{
		TotalFrames: len(frames),
	}

	aggressiveBits := 7
	info.SafeBitsPerFrame = aggressiveBits
	info.TotalCapacityBits = info.TotalFrames * aggressiveBits
	info.TotalCapacityBytes = info.TotalCapacityBits / 8

	return info
}

func displayCapacityResults(conservative, aggressive *HeaderEmbeddingInfo) {
	fmt.Printf("\n=== Header Embedding Capacity Analysis ===\n")
	fmt.Printf("Total frames: %d\n", conservative.TotalFrames)

	fmt.Printf("\n--- CONSERVATIVE APPROACH (Recommended) ---\n")
	fmt.Printf("Safe bits per frame: %d bits (Private, Copyright, Original)\n", conservative.SafeBitsPerFrame)
	fmt.Printf("Total available bits: %d bits\n", conservative.TotalCapacityBits)
	fmt.Printf("Total capacity: %s\n", formatBytes(conservative.TotalCapacityBytes))

	fmt.Printf("\n--- AGGRESSIVE APPROACH (Risky) ---\n")
	fmt.Printf("Usable bits per frame: %d bits (includes Protection, Padding, Emphasis)\n", aggressive.SafeBitsPerFrame)
	fmt.Printf("Total available bits: %d bits\n", aggressive.TotalCapacityBits)
	fmt.Printf("Total capacity: %s\n", formatBytes(aggressive.TotalCapacityBytes))

	fmt.Printf("\n=== Capacity Comparison ===\n")
	fmt.Printf("File Type              Conservative    Aggressive\n")
	fmt.Printf("Small text (1KB)       %s           %s\n",
		getCapacityEstimate(conservative.TotalCapacityBytes, 1024),
		getCapacityEstimate(aggressive.TotalCapacityBytes, 1024))
	fmt.Printf("Medium document (10KB) %s           %s\n",
		getCapacityEstimate(conservative.TotalCapacityBytes, 10*1024),
		getCapacityEstimate(aggressive.TotalCapacityBytes, 10*1024))
	fmt.Printf("Large document (100KB) %s           %s\n",
		getCapacityEstimate(conservative.TotalCapacityBytes, 100*1024),
		getCapacityEstimate(aggressive.TotalCapacityBytes, 100*1024))

	fmt.Printf("\n=== Recommendation ===\n")
	fmt.Printf("✓ CONSERVATIVE: Extremely safe, no audio quality impact\n")
	fmt.Printf("⚠ AGGRESSIVE: Higher capacity but may affect playback compatibility\n")
	fmt.Printf("🎯 For secret documents, conservative approach is recommended\n")
}

func formatBytes(bytes int) string {
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.2f MB (%d bytes)", float64(bytes)/1024/1024, bytes)
	} else if bytes >= 1024 {
		return fmt.Sprintf("%.2f KB (%d bytes)", float64(bytes)/1024, bytes)
	} else {
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func showFrameBreakdown(frames []*MP3FrameHeader, offsets []int, mp3Data []byte) {
	fmt.Printf("\n=== Sample Frame Analysis ===\n")

	for i, frame := range frames {
		offset := offsets[i]
		fmt.Printf("Frame %d (offset %d):\n", i+1, offset)
		fmt.Printf("  Size: %d bytes\n", frame.Size)
		fmt.Printf("  Bitrate: %d kbps\n", bitrateTable[frame.Bitrate])
		fmt.Printf("  Channel: %s\n", getChannelMode(frame.Channel))

		if offset+4 <= len(mp3Data) {
			fmt.Printf("  Header bytes: %02X %02X %02X %02X\n",
				mp3Data[offset], mp3Data[offset+1], mp3Data[offset+2], mp3Data[offset+3])

			fmt.Printf("  Safe bits: Private=%d, Copyright=%d, Original=%d\n",
				frame.Private, frame.Copyright, frame.Original)
		}
		fmt.Println()
	}
}

func getCapacityEstimate(available, needed int) string {
	if available >= needed {
		return "✓ YES"
	}
	return "✗ NO"
}

func getChannelMode(mode uint8) string {
	switch mode {
	case 0:
		return "Stereo"
	case 1:
		return "Joint Stereo"
	case 2:
		return "Dual Channel"
	case 3:
		return "Mono"
	default:
		return "Unknown"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
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
		return nil, fmt.Errorf("unsupported MP3 format (only MPEG1 Layer 3 supported)")
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
