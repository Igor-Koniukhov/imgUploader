package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"imageAploaderS3/models"
	"strings"
	"time"
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

func CalculateAge(birthdate time.Time) int {
	now := time.Now()
	years := now.Year() - birthdate.Year()
	if now.Month() < birthdate.Month() || (now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		years--
	}
	return years
}

func ParseBirthdate(birthdateStr string) (time.Time, error) {
	layout := "2006-01-02"
	return time.Parse(layout, birthdateStr)
}
