package blast

import (
	"github.com/ionrock/procs"
)

type ProcessPayload struct {
	content []byte
}

func NewProcessPayload(processName string) (*ProcessPayload, error) {
	process := procs.NewProcess(processName)
	if err := process.Run(); err != nil {
		return nil, err
	}
	content, err := process.Output()
	if err != nil {
		return nil, err
	}
	return &ProcessPayload{content: content}, nil
}

func (payload *ProcessPayload) Get() []byte {
	return payload.content
}
