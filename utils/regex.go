package utils

import (
	"fmt"
	"regexp"
)

const (
	// constants for regex pattern
	clientIDKey      = "clientID"
	batchNumberKey   = "batchNumber"
	messageNumberKey = "messageNumber"
	dataTypeKey      = "dataType"
	cityKey          = "city"
)

type RawMessageMetadata struct {
	ClientID      string
	BatchNumber   string
	MessageNumber string
	DataType      string
	City          string
}

// GetMetadataFromMessage extracts the clientID, batchNumber, messageNumber, dataType and city from the message
func GetMetadataFromMessage(message string) RawMessageMetadata {
	pattern := `^(?P<clientID>[^,]+),(?P<batchNumber>[^,]+),(?P<messageNumber>[^,]+),(?P<dataType>[^,]+),(?P<city>[^,]+)`
	regex := regexp.MustCompile(pattern)
	match := regex.FindStringSubmatch(message)

	// Extract the named groups from the match
	matches := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			if match[i] == "" {
				panic(fmt.Sprintf("missing metadata: %s", name))
			}
			matches[name] = match[i]
		}
	}

	return RawMessageMetadata{
		ClientID:      matches[clientIDKey],
		BatchNumber:   matches[batchNumberKey],
		MessageNumber: matches[messageNumberKey],
		DataType:      matches[dataTypeKey],
		City:          matches[cityKey],
	}
}

// GetFINMessage returns the FIN message within the message
func GetFINMessage(message string) string {
	regex := regexp.MustCompile(`(FIN-.*)`)
	matches := regex.FindStringSubmatch(message)
	if len(matches) != 2 {
		panic("invalid FIN message")
	}
	return matches[1]
}
