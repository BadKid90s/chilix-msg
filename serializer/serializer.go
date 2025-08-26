package serializer

type Serializer interface {
	Serialize(msg interface{}) ([]byte, error)
	Deserialize(data []byte, msg interface{}) error
}

var DefaultSerializer = &JSON{}
