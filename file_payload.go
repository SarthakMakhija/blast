package blast

import (
	"os"
)

type FilePayload struct {
	content []byte
}

func NewFilePayload(filePath string) (*FilePayload, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &FilePayload{content: content}, nil
}

func (payload *FilePayload) Get() []byte {
	return payload.content
}
