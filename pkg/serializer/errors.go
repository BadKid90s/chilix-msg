package serializer

import "errors"

var (
	ErrSerializationFailed   = errors.New("serialization failed")
	ErrDeserializationFailed = errors.New("deserialization failed")
	ErrInvalidPayloadType    = errors.New("invalid payload type for binary serializer")
)
