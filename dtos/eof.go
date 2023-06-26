package dtos

const eofType = "EOF"

// EOFData struct that it's send by the EOF Manager. Has two attributes that are in all domain entities.
// + Metadata: metadata added to the structure
type EOFData struct {
	Metadata Metadata `json:"metadata"`
}

func NewEOF(idempotencyKey string, city string, stage string, eofMessage string) *EOFData {
	return &EOFData{
		Metadata: NewMetadata(idempotencyKey, true, city, eofType, stage, eofMessage),
	}
}

func (eof EOFData) GetMetadata() Metadata {
	return eof.Metadata
}
