package core

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// turstileURL is the captcha verification API route.
const turnstileURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// turnstileResponse defines the Turnstile captcha JSON response.
type turnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// validateCaptcha valides a Turnstile captcha given an API secret,
// token, and remote ip address.
func validateCaptcha(secret, token, ip string) bool {
	payload := map[string]string{
		"secret":   secret,
		"response": token,
		"remoteip": ip,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	resp, err := http.Post(turnstileURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var tsResponse turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&tsResponse); err != nil {
		return false
	}
	if !tsResponse.Success {
		return false
	}
	return true
}
