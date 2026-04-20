package main

import (
	"os/exec"
	"strings"
	"testing"
)

func Test30579aa8(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("../ti", "./30579aa8.rb")

	output, _ := cmd.CombinedOutput()

	expectedOutput := `./30579aa8.rb:::8:::GPIO
./30579aa8.rb:::9:::Integer`

	if strings.TrimSpace(string(output)) != strings.TrimSpace(expectedOutput) {
		t.Errorf("Expected output: %s, but got: %s", expectedOutput, string(output))
	}
}
