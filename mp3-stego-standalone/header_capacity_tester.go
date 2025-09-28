package main

import (
	"fmt"
	"io"
	"os"
)

// MP3FrameHeader represents an MP3 frame header
type MP3FrameHeader struct {
	Sync       uint16 // Frame sync (11 bits)
	Version    uint8  // MPEG Audio version
	Layer      uint8  // Layer description
	Protection uint8  // Protection bit
	Bitrate    uint8  // Bitrate index
	SampleRate uint8  // Sampling rate frequency index
	Padding    uint8  // Padding bit
	Private    uint8  // Private bit
	Channel    uint8  // Channel Mode
	ModeExt    uint8  // Mode extension
	Copyright  uint8  // Copyright
	Original   uint8  // Original
	Emphasis   uint8  // Emphasis
	Size       int    // Frame size in bytes
}

// Bitrate table for MPEG1 Layer 3 (most common)
var bitrateTable = []int{
	0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0,
}

// Sample rate table for MPEG1
var sampleRateTable = []int{
	44100, 48000, 32000, 0,
}

// HeaderEmbeddingInfo contains information about embedding capacity in headers
type HeaderEmbeddingInfo struct {
	TotalFrames       int
	UsableHeaderBits  int
	SafeBitsPerFrame  int
	TotalCapacityBits int
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

// analyzeMP3HeaderCapacity analyzes an MP3 file and reports header embedding capacity
func analyzeMP3HeaderCapacity(mp3Path string) error {
	// Read MP3 file
	mp3Data, err := readFile(mp3Path)
	if err != nil {
		return fmt.Errorf("failed to read MP3 file: %v", err)
	}

	fmt.Printf("=== MP3 Header Capacity Analysis ===\n")
	fmt.Printf("File: %s\n", mp3Path)
	fmt.Printf("File size: %d bytes (%.2f MB)\n", len(mp3Data), float64(len(mp3Data))/1024/1024)

	// Skip ID3 tag if present
	dataStart := skipID3Tag(mp3Data)
	if dataStart > 0 {
		fmt.Printf("ID3 tag size: %d bytes (skipped)\n", dataStart)
	}

	// Parse MP3 frames
	frames, offsets, err := findMP3Frames(mp3Data[dataStart:])
	if err != nil {
		return fmt.Errorf("failed to parse MP3 frames: %v", err)
	}

	if len(frames) == 0 {
		return fmt.Errorf("no valid MP3 frames found")
	}

	// Adjust offsets for ID3 tag skip
	for i := range offsets {
		offsets[i] += dataStart
	}

	fmt.Printf("Total MP3 frames found: %d\n", len(frames))

	// Analyze header embedding capacity (conservative)
	conservativeInfo := analyzeHeaderEmbeddingCapacity(frames, offsets, mp3Data)
	
	// Analyze aggressive capacity
	aggressiveInfo := analyzeAggressiveCapacity(frames)
	
	// Display detailed results
	displayCapacityResults(conservativeInfo, aggressiveInfo)

	// Show frame breakdown for first few frames
	showFrameBreakdown(frames[:min(5, len(frames))], offsets[:min(5, len(offsets))], mp3Data)

	return nil
}

// analyzeHeaderEmbeddingCapacity calculates how much data can be embedded in headers
func analyzeHeaderEmbeddingCapacity(frames []*MP3FrameHeader, offsets []int, mp3Data []byte) *HeaderEmbeddingInfo {
	info := &HeaderEmbeddingInfo{
		TotalFrames: len(frames),
	}

	// For each frame header (4 bytes), we can potentially use some bits
	// We need to be very careful about which bits we can safely modify
	
	// Safe bits in MP3 frame header:
	// Byte 0: Frame sync (0xFF) - DO NOT MODIFY
	// Byte 1[7:5]: Frame sync continuation (111) - DO NOT MODIFY
	// Byte 1[4:3]: Version - DO NOT MODIFY
	// Byte 1[2:1]: Layer - DO NOT MODIFY  
	// Byte 1[0]: Protection bit - POTENTIALLY SAFE (CRC on/off)
	// Byte 2[7:4]: Bitrate - DO NOT MODIFY
	// Byte 2[3:2]: Sample rate - DO NOT MODIFY
	// Byte 2[1]: Padding - POTENTIALLY SAFE
	// Byte 2[0]: Private bit - SAFE (user defined)
	// Byte 3[7:6]: Channel mode - DO NOT MODIFY
	// Byte 3[5:4]: Mode extension - POTENTIALLY SAFE (for joint stereo)
	// Byte 3[3]: Copyright - SAFE (metadata)
	// Byte 3[2]: Original - SAFE (metadata)
	// Byte 3[1:0]: Emphasis - POTENTIALLY SAFE (rarely used)

	// Conservative approach: Only use clearly safe bits
	// - Private bit (1 bit per frame)
	// - Copyright bit (1 bit per frame) 
	// - Original bit (1 bit per frame)
	// Total: 3 bits per frame header

	conservativeBits := 3 // private, copyright, original bits
	info.SafeBitsPerFrame = conservativeBits
	info.TotalCapacityBits = info.TotalFrames * conservativeBits
	info.TotalCapacityBytes = info.TotalCapacityBits / 8

	return info
}

// analyzeAggressiveCapacity calculates capacity with more aggressive bit usage
func analyzeAggressiveCapacity(frames []*MP3FrameHeader) *HeaderEmbeddingInfo {
	info := &HeaderEmbeddingInfo{
		TotalFrames: len(frames),
	}
	
	// More aggressive approach could include:
	// - Private bit (1 bit) - SAFE
	// - Copyright bit (1 bit) - SAFE  
	// - Original bit (1 bit) - SAFE
	// - Protection bit (1 bit) - RISKY but often unused
	// - Padding bit (1 bit) - RISKY but sometimes safe
	// - Emphasis bits (2 bits) - RISKY but rarely used
	// Total: up to 7 bits per frame header
	
	aggressiveBits := 7 // All potentially usable bits
	info.SafeBitsPerFrame = aggressiveBits
	info.TotalCapacityBits = info.TotalFrames * aggressiveBits
	info.TotalCapacityBytes = info.TotalCapacityBits / 8

	return info
}

// displayCapacityResults shows the embedding capacity analysis
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

// formatBytes formats byte count into human-readable string
func formatBytes(bytes int) string {
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.2f MB (%d bytes)", float64(bytes)/1024/1024, bytes)
	} else if bytes >= 1024 {
		return fmt.Sprintf("%.2f KB (%d bytes)", float64(bytes)/1024, bytes)
	} else {
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// showFrameBreakdown shows detailed info for sample frames
func showFrameBreakdown(frames []*MP3FrameHeader, offsets []int, mp3Data []byte) {
	fmt.Printf("\n=== Sample Frame Analysis ===\n")
	
	for i, frame := range frames {
		offset := offsets[i]
		fmt.Printf("Frame %d (offset %d):\n", i+1, offset)
		fmt.Printf("  Size: %d bytes\n", frame.Size)
		fmt.Printf("  Bitrate: %d kbps\n", bitrateTable[frame.Bitrate])
		fmt.Printf("  Channel: %s\n", getChannelMode(frame.Channel))
		
		// Show header bytes
		if offset+4 <= len(mp3Data) {
			fmt.Printf("  Header bytes: %02X %02X %02X %02X\n", 
				mp3Data[offset], mp3Data[offset+1], mp3Data[offset+2], mp3Data[offset+3])
			
			// Show which bits we could potentially modify
			fmt.Printf("  Safe bits: Private=%d, Copyright=%d, Original=%d\n",
				frame.Private, frame.Copyright, frame.Original)
		}
		fmt.Println()
	}
}

// getCapacityEstimate returns whether a file of given size can fit
func getCapacityEstimate(available, needed int) string {
	if available >= needed {
		return "✓ YES"
	}
	return "✗ NO"
}

// getChannelMode returns human-readable channel mode
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// readFile reads a file and returns its contents
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

// skipID3Tag skips ID3v2 tag if present
func skipID3Tag(data []byte) int {
	if len(data) < 10 {
		return 0
	}
	
	// Check for ID3v2 tag
	if data[0] == 'I' && data[1] == 'D' && data[2] == '3' {
		// Get tag size (synchsafe integer)
		size := int(data[6])<<21 | int(data[7])<<14 | int(data[8])<<7 | int(data[9])
		return 10 + size // Header + tag size
	}
	
	return 0
}

// parseMP3Frame parses an MP3 frame header and returns frame information
func parseMP3Frame(data []byte, offset int) (*MP3FrameHeader, error) {
	if len(data) < offset+4 {
		return nil, fmt.Errorf("not enough data for frame header")
	}

	// Check for frame sync (11 bits of 1s)
	if data[offset] != 0xFF || (data[offset+1]&0xE0) != 0xE0 {
		return nil, fmt.Errorf("invalid frame sync")
	}

	header := &MP3FrameHeader{}
	
	// Parse header bytes
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

	// Calculate frame size for MPEG1 Layer 3
	if header.Version == 3 && header.Layer == 1 { // MPEG1 Layer 3
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

// findMP3Frames finds all MP3 frames in the data
func findMP3Frames(data []byte) ([]*MP3FrameHeader, []int, error) {
	var frames []*MP3FrameHeader
	var offsets []int
	
	i := 0
	for i < len(data)-4 {
		// Look for frame sync
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