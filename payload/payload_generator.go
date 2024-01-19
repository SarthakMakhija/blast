package payload

type ConstantPayloadGenerator struct {
	payload []byte
}

func NewConstantPayloadGenerator(payload []byte) ConstantPayloadGenerator {
	return ConstantPayloadGenerator{
		payload: payload,
	}
}

func (generator ConstantPayloadGenerator) Generate(_ uint) []byte {
	return generator.payload
}
