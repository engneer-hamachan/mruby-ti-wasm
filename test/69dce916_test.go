package main

import (
	"os/exec"
	"strings"
	"testing"
)

func Test69dce916(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("../ti", "./69dce916.rb")

	output, _ := cmd.CombinedOutput()

	expectedOutput := "./69dce916.rb:::1:::type mismatch: expected Union<Integer Float>, but got Hash for Integer.+"

	if strings.TrimSpace(string(output)) != strings.TrimSpace(expectedOutput) {
		t.Errorf("Expected output: %s, but got: %s", expectedOutput, string(output))
	}
}
