package helper

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GenerateRandomString generates a random string of a given length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateRandomEmail generates a random email with the domain @gmail.com
func GenerateRandomEmail() string {
	usernameLength := rand.Intn(10) + 5 // Username length between 5 and 15 characters
	username := GenerateRandomString(usernameLength)
	email := fmt.Sprintf("%s@gmail.com", strings.ToLower(username))
	return email
}

// GenerateRandomFloat generates a random float64 within a specified range
func GenerateRandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// GenerateRandomLatLong generates a random latitude and longitude
func GenerateRandomLatLong() (latitude float64, longitude float64) {
	rand.Seed(time.Now().UnixNano())
	latitude = GenerateRandomFloat(-90.0, 90.0)
	longitude = GenerateRandomFloat(-180.0, 180.0)
	return latitude, longitude
}