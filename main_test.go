package main_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Gulliby-development/fpbits/internal/cli"
)

func TestREADMEExamplesMatchCLI(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	// Markdown source may use CRLF, but the CLI output being compared remains untouched.
	markdown := strings.ReplaceAll(string(readme), "\r\n", "\n")
	blocks, err := readFencedBlocks(markdown)
	if err != nil {
		t.Fatal(err)
	}

	usage, err := uniqueBlockWithPrefix(blocks, "usage: fpbits decode FORMAT HEX...\n", "usage")
	if err != nil {
		t.Fatal(err)
	}
	rangeF16, err := uniqueBlockWithPrefix(blocks, "format: f16\n", "range f16")
	if err != nil {
		t.Fatal(err)
	}

	assertDocumentExample(t, []string{"--help"}, usage)
	assertDocumentExample(t, []string{"range", "f16"}, rangeF16)
}

func readFencedBlocks(markdown string) ([]string, error) {
	var blocks []string
	var body strings.Builder
	inFence := false

	for _, line := range strings.SplitAfter(markdown, "\n") {
		marker := strings.TrimSpace(strings.TrimSuffix(line, "\n"))
		if !inFence && strings.HasPrefix(marker, "```") {
			inFence = true
			body.Reset()
			continue
		}
		if inFence && marker == "```" {
			blocks = append(blocks, body.String())
			inFence = false
			continue
		}
		if inFence {
			body.WriteString(line)
		}
	}

	if inFence {
		return nil, errors.New("README.md contains an unterminated fenced code block")
	}
	return blocks, nil
}

func uniqueBlockWithPrefix(blocks []string, prefix, label string) (string, error) {
	var matches []string
	for _, block := range blocks {
		if strings.HasPrefix(block, prefix) {
			matches = append(matches, block)
		}
	}
	switch len(matches) {
	case 0:
		return "", errors.New("README.md is missing the " + label + " output fence")
	case 1:
		return matches[0], nil
	default:
		return "", errors.New("README.md contains duplicate " + label + " output fences")
	}
}

func assertDocumentExample(t *testing.T, args []string, expected string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	status := cli.Run(args, strings.NewReader(""), &stdout, &stderr)
	if status != 0 {
		t.Errorf("cli.Run(%q) status = %d, want 0 (stderr %q)", args, status, stderr.String())
	}
	if got := stdout.String(); got != expected {
		t.Errorf("cli.Run(%q) stdout = %q, want README fence %q", args, got, expected)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("cli.Run(%q) stderr = %q, want empty", args, got)
	}
}

func TestMalformedUTF8OnStandardInput(t *testing.T) {
	executableName := "fpbits"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executable := filepath.Join(t.TempDir(), executableName)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	build := exec.CommandContext(ctx, "go", "build", "-o", executable, ".")
	var buildStdout, buildStderr bytes.Buffer
	build.Stdout = &buildStdout
	build.Stderr = &buildStderr
	if err := build.Run(); err != nil {
		t.Fatalf("build fpbits: %v\nstdout: %s\nstderr: %s", err, buildStdout.String(), buildStderr.String())
	}

	command := exec.CommandContext(ctx, executable, "decode", "f16", "-")
	command.Stdin = bytes.NewReader([]byte{0x33, 0x63, 0x30, 0x30, 0x20, 0xff, 0x0a})
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()

	var exitError *exec.ExitError
	if err == nil {
		t.Fatalf("fpbits exited successfully, want status 2 (stdout %q, stderr %q)", stdout.String(), stderr.String())
	}
	if !errors.As(err, &exitError) {
		t.Fatalf("could not run fpbits: %v", err)
	}
	if code := exitError.ExitCode(); code != 2 {
		t.Errorf("fpbits exit status = %d, want 2 (stderr %q)", code, stderr.String())
	}
	if got := stdout.Bytes(); len(got) != 0 {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.Bytes(), []byte("fpbits: <stdin>: input is not valid UTF-8\n"); !bytes.Equal(got, want) {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}
