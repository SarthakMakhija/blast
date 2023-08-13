package blast

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadsTheFile(t *testing.T) {
	file, err := os.CreateTemp(".", "file_payload")
	assert.Nil(t, err)

	defer os.Remove(file.Name())

	_, err = file.Write([]byte("sample test content"))
	assert.Nil(t, err)

	filePayload, err := NewFilePayload(file.Name())
	assert.Nil(t, err)

	assert.Equal(t, "sample test content", string(filePayload.Get()))
}

func TestReadsTheNonExistentFile(t *testing.T) {
	_, err := NewFilePayload("non-existing")
	assert.Error(t, err)
}
