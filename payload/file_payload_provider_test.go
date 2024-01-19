package payload

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadsTheFile(t *testing.T) {
	file, err := os.CreateTemp(".", "file_payload")
	assert.Nil(t, err)

	defer func(name string) {
		_ = os.Remove(name)
	}(file.Name())

	_, err = file.Write([]byte("sample test content"))
	assert.Nil(t, err)

	filePayloadProvider, err := NewFilePayloadProvider(file.Name())
	assert.Nil(t, err)

	assert.Equal(t, "sample test content", string(filePayloadProvider.Get()))
}

func TestReadsTheNonExistentFile(t *testing.T) {
	_, err := NewFilePayloadProvider("non-existing")
	assert.Error(t, err)
}
