package serializer

import "encoding/json"

// JSON JSON序列化器
type JSON struct{}

func (j *JSON) Serialize(msg interface{}) ([]byte, error) {
	return json.Marshal(msg)
}

func (j *JSON) Deserialize(data []byte, msg interface{}) error {
	return json.Unmarshal(data, msg)
}
