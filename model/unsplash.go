package model

import "time"

type Image struct {
	ImageUrl             string        `json:"image_url"`
	PhotographerName     string        `json:"photographer_name"`
	PhotographerUsername string        `json:"photographer_username"`
	Updated              time.Time     `json:"updated_at"`
	ExpirationDuration   time.Duration `json:"expiration_duration"`
}
