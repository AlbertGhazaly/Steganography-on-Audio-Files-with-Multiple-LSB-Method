package models

type EmbedRequest struct {
	Key               string `json:"key"`
	UseEncryption     bool   `json:"use_encryption"`
	UseKeyForPosition bool   `json:"use_key_for_position"`
	LSBBits           int    `json:"lsb_bits"`
}

type ExtractRequest struct {
	Key               string `json:"key"`
	UseEncryption     bool   `json:"use_encryption"`
	UseKeyForPosition bool   `json:"use_key_for_position"`
	LSBBits           int    `json:"lsb_bits"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}
