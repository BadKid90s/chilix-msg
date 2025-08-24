// pkg/codec/codec_test.go

package codec_test

import (
	"bytes"
	"testing"

	"github.com/BadKid90s/chilix-msg/pkg/codec"
	"github.com/BadKid90s/chilix-msg/pkg/serializer"
	"github.com/stretchr/testify/assert"
)

func TestCodecRoundTrip(t *testing.T) {
	serializer := serializer.DefaultSerializer
	codec := codec.NewLengthPrefixCodec(serializer)

	// 创建缓冲区
	buf := &bytes.Buffer{}

	// 编码消息
	msgType := "test"
	payload := map[string]string{"key": "value"}
	requestID := uint64(12345)
	err := codec.Encode(buf, msgType, payload, requestID)
	assert.NoError(t, err)

	// 解码消息
	decodedMsgType, decodedPayload, decodedRequestID, err := codec.Decode(buf)
	assert.NoError(t, err)

	// 验证结果
	assert.Equal(t, msgType, decodedMsgType)
	assert.Equal(t, requestID, decodedRequestID)

	// 验证负载
	var decodedData map[string]string
	err = serializer.Deserialize(decodedPayload, &decodedData)
	assert.NoError(t, err)
	assert.Equal(t, "value", decodedData["key"])
}
