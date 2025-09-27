package stego

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
)

// PaperLSBSteganography implements a paper-style LSB method with:
// - random start position (optionally keyed)
// - start and end signatures that encode the number of bits used
// - 1–4 LSBs per carrier byte
// Note: This operates over MP3 bytes (after a header skip) rather than decoded PCM samples.
// It is a pragmatic adaptation to match the paper's flow (random start + signatures + multi-bit LSB).
type PaperLSBSteganography struct {
	headerSize int
}

func NewPaperLSBSteganography() *PaperLSBSteganography {
	return &PaperLSBSteganography{headerSize: 1024}
}

// signature patterns as described by the paper examples.
// A = 10101010101010 (14 bits), B = 01010101010101 (14 bits)
func signatureBitsA() []byte {
	// 14 bits alternating starting with 1
	bits := make([]byte, 14)
	for i := 0; i < 14; i++ {
		if i%2 == 0 {
			bits[i] = 1
		} else {
			bits[i] = 0
		}
	}
	return bits
}

func signatureBitsB() []byte {
	// 14 bits alternating starting with 0
	bits := make([]byte, 14)
	for i := 0; i < 14; i++ {
		if i%2 == 0 {
			bits[i] = 0
		} else {
			bits[i] = 1
		}
	}
	return bits
}

// getSignature returns the concatenated signature bits for the given LSB count.
// 1-bit: A
// 2-bit: B
// 3-bit: A + B
// 4-bit: B + A
func getSignature(lsbBits int) []byte {
	a := signatureBitsA()
	b := signatureBitsB()
	switch lsbBits {
	case 1:
		return append([]byte{}, a...)
	case 2:
		return append([]byte{}, b...)
	case 3:
		out := make([]byte, 0, len(a)+len(b))
		out = append(out, a...)
		out = append(out, b...)
		return out
	case 4:
		out := make([]byte, 0, len(a)+len(b))
		out = append(out, b...)
		out = append(out, a...)
		return out
	default:
		return nil
	}
}

// toBitStream converts a byte slice into a bit stream (MSB-first per byte).
func toBitStream(data []byte) []byte {
	bits := make([]byte, 0, len(data)*8)
	for _, by := range data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (by>>uint(i))&1)
		}
	}
	return bits
}

// fromBitStream converts a bit stream (MSB first) into bytes; ignores extra trailing bits <8.
func fromBitStream(bits []byte) []byte {
	if len(bits) == 0 {
		return []byte{}
	}
	outLen := len(bits) / 8
	out := make([]byte, outLen)
	for i := 0; i < outLen; i++ {
		var by byte = 0
		for j := 0; j < 8; j++ {
			by <<= 1
			by |= bits[i*8+j] & 1
		}
		out[i] = by
	}
	return out
}

// combineK merges the next k bits (MSB-first) into a value 0..(2^k-1).
func combineK(bits []byte, start, k int) (byte, int) {
	var v byte = 0
	for i := 0; i < k; i++ {
		v <<= 1
		if start+i < len(bits) {
			v |= bits[start+i] & 1
		}
	}
	return v, start + k
}

// searchPattern searches for pattern in bits and returns the bit index of the first match or -1.
func searchPattern(bits, pattern []byte) int {
	if len(pattern) == 0 || len(bits) < len(pattern) {
		return -1
	}
	// naive scan; acceptable for typical sizes
	for i := 0; i <= len(bits)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if bits[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// computeSeed derives a deterministic seed from the provided key.
func computeSeed(key string) int64 {
	if key == "" {
		// fallback seed from random source
		// rand.Seed will be set by caller; return a default value
		return int64(0x6a09e667f3bcc909) // valid int64 default seed
	}
	sum := sha256.Sum256([]byte(key))
	// take first 8 bytes little endian as seed
	return int64(binary.LittleEndian.Uint64(sum[:8]))
}

// carriersNeeded returns the number of carrier bytes needed for a given number of message bits
// when embedding k bits per carrier byte.
func carriersNeeded(totalBits, k int) int {
	return int(math.Ceil(float64(totalBits) / float64(k)))
}

// PaperCalculateCapacity returns the maximum payload bytes that can be embedded for a given MP3 size and k,
// accounting for start and end signatures.
func PaperCalculateCapacity(mp3Size int, k int) int {
	if mp3Size <= 1024 || k < 1 || k > 4 {
		return 0
	}
	signatureLenBits := len(getSignature(k))
	totalCarrier := mp3Size - 1024
	// total available bits in LSB stream
	totalBits := totalCarrier * k
	overheadBits := 2 * signatureLenBits // start + end
	if totalBits <= overheadBits {
		return 0
	}
	payloadBits := totalBits - overheadBits
	return payloadBits / 8
}

// EmbedMessagePaper embeds message using the paper-style method.
// If useKeyForPosition is true, the random start is derived from key; otherwise, it's pseudo-random.
func (p *PaperLSBSteganography) EmbedMessagePaper(mp3Data, message []byte, k int, key string, useKeyForPosition bool) ([]byte, error) {
	if k < 1 || k > 4 {
		return nil, ErrInvalidBitCount
	}
	if len(mp3Data) <= p.headerSize {
		return nil, ErrInvalidMP3Format
	}

	// Compose bitstream = startSig + messageBits + endSig
	startSig := getSignature(k)
	endSig := getSignature(k)
	msgBits := toBitStream(message)

	totalBits := len(startSig) + len(msgBits) + len(endSig)
	// compute carriers needed and capacity
	carriers := carriersNeeded(totalBits, k)
	availableCarriers := len(mp3Data) - p.headerSize

	// need room for 200-byte guard (paper uses 200 samples); apply same in bytes
	guard := 200
	if carriers+guard > availableCarriers {
		return nil, ErrInsufficientCapacity
	}

	// Determine random start (in carrier bytes after header)
	espace := availableCarriers - carriers - guard
	if espace <= 0 {
		return nil, ErrInsufficientCapacity
	}

	var rng *rand.Rand
	if useKeyForPosition {
		seed := computeSeed(key)
		rng = rand.New(rand.NewSource(seed))
	} else {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	// choose within [guard, guard + floor(espace/2))
	startOffset := guard
	if espace/2 > 0 {
		startOffset = guard + rng.Intn(espace/2)
	}

	// Build embedding over a copy
	out := make([]byte, len(mp3Data))
	copy(out, mp3Data)

	carrier := out[p.headerSize:]

	// iterator over bitstream
	bitStream := make([]byte, 0, totalBits)
	bitStream = append(bitStream, startSig...)
	bitStream = append(bitStream, msgBits...)
	bitStream = append(bitStream, endSig...)

	// embed k bits per carrier byte
	bi := 0
	for i := 0; i < carriers; i++ {
		if startOffset+i >= len(carrier) {
			break
		}
		val, next := combineK(bitStream, bi, k)
		bi = next
		// clear k LSBs and set
		mask := byte((1 << k) - 1)
		carrier[startOffset+i] = (carrier[startOffset+i] & ^mask) | (val & mask)
		if bi >= len(bitStream) {
			break
		}
	}

	// sanity check
	if bi < len(bitStream) {
		return nil, errors.New("internal embedding overflow: insufficient carriers")
	}

	return out, nil
}

// ExtractMessagePaper extracts message embedded using the paper-style method.
// It scans the entire LSB stream for the start signature, then collects bits until the end signature.
func (p *PaperLSBSteganography) ExtractMessagePaper(mp3Data []byte, k int) ([]byte, error) {
	if k < 1 || k > 4 {
		return nil, ErrInvalidBitCount
	}
	if len(mp3Data) <= p.headerSize {
		return nil, ErrInvalidMP3Format
	}

	carrier := mp3Data[p.headerSize:]
	// Build LSB bitstream from entire carrier: k bits per byte
	bits := make([]byte, 0, len(carrier)*k)
	mask := byte((1 << k) - 1)
	for i := 0; i < len(carrier); i++ {
		v := carrier[i] & mask
		// expand v into k bits, MSB-first
		for j := k - 1; j >= 0; j-- {
			bits = append(bits, (v>>uint(j))&1)
		}
	}

	sig := getSignature(k)
	if len(sig) == 0 {
		return nil, ErrInvalidBitCount
	}

	startIdx := searchPattern(bits, sig)
	if startIdx < 0 {
		return nil, errors.New("start signature not found")
	}

	// search for end signature after the start
	// To reduce false positives for k=3,4 where signature is longer, this should be reasonably unique.
	afterStart := startIdx + len(sig)
	if afterStart >= len(bits) {
		return nil, ErrInvalidMP3Format
	}
	endIdx := searchPattern(bits[afterStart:], sig)
	if endIdx < 0 {
		return nil, errors.New("end signature not found")
	}
	endIdx += afterStart

	payloadBits := bits[afterStart:endIdx]
	// convert to bytes
	out := fromBitStream(payloadBits)
	return out, nil
}
