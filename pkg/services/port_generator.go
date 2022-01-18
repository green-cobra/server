package services

import (
	"math/rand"
	"net"
	"strconv"
	"time"
)

func GenerateOpenedPortNumber(min, max int) int {
	randPort := func() int {
		return min + rand.Intn(max-min)
	}

	port := randPort()
	for IsPortTaken(port) {
		port = randPort()
	}
	return port
}

func IsPortTaken(port int) bool {
	timeout := time.Second
	conn, _ := net.DialTimeout("tcp", net.JoinHostPort("localhost", strconv.Itoa(port)), timeout)
	if conn != nil {
		defer conn.Close()
		return true
	}

	return false
}
