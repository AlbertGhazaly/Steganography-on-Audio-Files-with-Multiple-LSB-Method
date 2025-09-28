package crypto

func VigenereEncrypt(data []byte, key string) []byte {
	if len(key) == 0 {
		return data
	}

	result := make([]byte, len(data))
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	for i := 0; i < len(data); i++ {
		keyByte := keyBytes[i%keyLen]
		result[i] = byte((uint(data[i]) + uint(keyByte)) % 256)
	}

	return result
}
