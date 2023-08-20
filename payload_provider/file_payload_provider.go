package payloadprovider

import (
	"os"
)

// FilePayloadProvider reads the payload from a file.
// Payload returned by FilePayloadProvider is used in load generation.
type FilePayloadProvider struct {
	content []byte
}

// NewFilePayloadProvider returns the FilePayloadProvider from the provided filePath.
func NewFilePayloadProvider(filePath string) (*FilePayloadProvider, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &FilePayloadProvider{content: content}, nil
}

// Get returns the file content.
// The entire content of the file is read and returned as byte slice.
func (payloadProvider *FilePayloadProvider) Get() []byte {
	return payloadProvider.content
}
