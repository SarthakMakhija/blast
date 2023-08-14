package payloadprovider

import (
	"os"
)

type FilePayloadProvider struct {
	content []byte
}

func NewFilePayloadProvider(filePath string) (*FilePayloadProvider, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &FilePayloadProvider{content: content}, nil
}

func (payloadProvider *FilePayloadProvider) Get() []byte {
	return payloadProvider.content
}
