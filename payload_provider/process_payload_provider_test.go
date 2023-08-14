package payloadprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadsAProcessOutput(t *testing.T) {
	file, err := os.CreateTemp(".", "test_file")
	assert.Nil(t, err)

	defer os.Remove(file.Name())

	processPayload, err := NewProcessPayloadProvider("ls | grep test_file*")

	assert.Equal(t, filepath.Base(file.Name()), string(processPayload.Get()))
}

func TestReadsANonExisitngProcess(t *testing.T) {
	_, err := NewProcessPayloadProvider("non-existing")

	assert.Error(t, err)
}