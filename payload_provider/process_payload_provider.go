package payloadprovider

import (
	"github.com/ionrock/procs"
)

type ProcessPayloadProvider struct {
	content []byte
}

func NewProcessPayloadProvider(processName string) (*ProcessPayloadProvider, error) {
	process := procs.NewProcess(processName)
	if err := process.Run(); err != nil {
		return nil, err
	}
	content, err := process.Output()
	if err != nil {
		return nil, err
	}
	return &ProcessPayloadProvider{content: content}, nil
}

func (payloadProvider *ProcessPayloadProvider) Get() []byte {
	return payloadProvider.content
}
