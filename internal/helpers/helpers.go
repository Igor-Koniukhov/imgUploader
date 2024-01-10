package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"imageAploaderS3/models"
	"strings"
)

func GetJWTPayloadData(jwt string) *models.JWTPayload {

	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		fmt.Println("Invalid JWT")
		return nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Println("Error in decoding base64:", err)
		return nil
	}

	var payloadData models.JWTPayload
	if err := json.Unmarshal(payload, &payloadData); err != nil {
		fmt.Println("Error in unmarshalling JSON:", err)
		return nil
	}

	return &payloadData
}
