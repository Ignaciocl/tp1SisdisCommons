package dtos

// Metadata this struct will contain extra information about the data that travels in our system
// + IdempotencyKey: idempotency key of the data which this metadata is associated
// + EOF: attribute to know if the given data corresponds to an EOF message
// + City: city which belongs the data
// + DataType: this field helps us to recognize in different stages what type of data is, possible values are: trips, stations and weather
// + Stage: stage were the Metadata was constructed
// + Message: message with extra information
type Metadata struct {
	IdempotencyKey string `json:"idempotency_key"`
	EOF            bool   `json:"eof"`
	City           string `json:"city"`
	DataType       string `json:"data_type"`
	Stage          string `json:"stage"`
	Message        string `json:"message"`
}

func NewMetadata(idempotencyKey string, eof bool, city string, dataType string, stage string, message string) Metadata {
	return Metadata{
		IdempotencyKey: idempotencyKey,
		EOF:            eof,
		City:           city,
		DataType:       dataType,
		Stage:          stage,
		Message:        message,
	}
}

func (m Metadata) GetIdempotencyKey() string {
	return m.IdempotencyKey
}

func (m Metadata) IsEOF() bool {
	return m.EOF
}

func (m Metadata) GetDataType() string {
	return m.DataType
}

func (m Metadata) GetCity() string {
	return m.City
}

func (m Metadata) GetStage() string {
	return m.Stage
}

func (m Metadata) GetMessage() string {
	return m.Message
}
