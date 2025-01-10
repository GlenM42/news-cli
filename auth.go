package main

import (
	"encoding/json"
	"os"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// AuthorizedUser represents a single user entry from authorized_keys.json
type AuthorizedUser struct {
	Username  string `json:"username"`
	PublicKey string `json:"publicKey"`
}

// AuthorizedUsers is a wrapper for the list of users in authorized_keys.json
type AuthorizedUsers struct {
	Users []AuthorizedUser `json:"users"`
}

// authorizedUsers is a global slice where we store all loaded users.
var authorizedUsers []AuthorizedUser

// loadAuthorizedKeys reads the JSON file containing user data and stores it in authorizedUsers.
func loadAuthorizedKeys(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var result AuthorizedUsers
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	authorizedUsers = result.Users
	return nil
}

// publicKeyAuth checks if the provided public key matches one of the keys in authorizedUsers.
func publicKeyAuth(_ ssh.Context, key ssh.PublicKey) bool {
	log.Info("Attempting public-key authentication")

	for _, user := range authorizedUsers {
		parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(user.PublicKey))
		if err != nil {
			continue // skip any malformed keys
		}
		if ssh.KeysEqual(key, parsed) {
			log.Info("Public-key authentication successful")
			return true
		}
	}
	log.Warn("Public-key authentication failed")
	return false
}

// keyboardInteractiveAuth prompts the user for two answers and validates them against environment variables.
func keyboardInteractiveAuth(_ ssh.Context, challenger gossh.KeyboardInteractiveChallenge) bool {
	log.Info("Attempting keyboard-interactive authentication")
	answers, err := challenger(
		"Authentication Challenge",
		"",
		[]string{
			"♦ Which editor is best (vim/emacs)? ",
			"♦ What is the name of the author of this app? ",
		},
		[]bool{true, true}, // Both prompts require answers
	)
	if err != nil || len(answers) != 2 {
		log.Warn("Keyboard-interactive authentication failed")
		return false
	}

	// Check correctness using environment variables
	correctAnswers := map[int]string{
		1: os.Getenv("QUESTION_1"),
		2: os.Getenv("QUESTION_2"),
	}
	if answers[0] == correctAnswers[1] && answers[1] == correctAnswers[2] {
		log.Info("Keyboard-interactive authentication successful")
		return true
	}

	log.Warn("Keyboard-interactive authentication failed")
	return false
}
