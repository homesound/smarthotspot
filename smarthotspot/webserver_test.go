package main

import "testing"

func TestWebserver(t *testing.T) {
	sPath := ".."
	serverPath = &sPath
	StartWebServer(nil, 31223)
}
