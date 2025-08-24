package serializer

// BinarySerializer 二进制序列化器
type BinarySerializer struct{}

func (b *BinarySerializer) Serialize(msg interface{}) ([]byte, error) {
	// 检查是否是字节切片
	if data, ok := msg.([]byte); ok {
		return data, nil
	}

	// 检查是否是字节切片指针
	if dataPtr, ok := msg.(*[]byte); ok {
		return *dataPtr, nil
	}

	return nil, ErrInvalidPayloadType
}

func (b *BinarySerializer) Deserialize(data []byte, msg interface{}) error {
	// 检查目标是否是字节切片
	if target, ok := msg.(*[]byte); ok {
		*target = data
		return nil
	}

	return ErrInvalidPayloadType
}
