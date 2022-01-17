package services

import (
	"math/rand"
	"strings"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func GetTunnelNameFromHost(host string) string {
	parts := strings.Split(host, ".")

	return parts[0]
}

func GenerateTunnelName() string {
	return randSeq(10)
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
