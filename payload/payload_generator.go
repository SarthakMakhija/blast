package payload

// ConstantPayloadGenerator provides a constant payload to all the workers for sending the payload.
type ConstantPayloadGenerator struct {
	payload []byte
}

// NewConstantPayloadGenerator creates a new instance of ConstantPayloadGenerator.
func NewConstantPayloadGenerator(payload []byte) ConstantPayloadGenerator {
	return ConstantPayloadGenerator{
		payload: payload,
	}
}

// Generate generates (/returns) the same payload for each request.
func (generator ConstantPayloadGenerator) Generate(_ uint) []byte {
	return generator.payload
}
