package benchmarks

import (
	"bytes"
	"testing"

	"github.com/BadKid90s/chilix-msg/codec"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/stretchr/testify/assert"
)

// 协议开销分析测试套件
// 用于验证协议头部开销，监控协议优化效果

// TestProtocolOverhead_V0_0_2 测试v0.0.2协议的开销
func TestProtocolOverhead_V0_0_2(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	tests := []struct {
		name             string
		msgType          string
		payload          interface{}
		expectedOverhead int
	}{
		{
			name:             "Small_Message",
			msgType:          "test",
			payload:          "hello",
			expectedOverhead: 15 + len("test"), // 15字节固定头部 + 消息类型长度
		},
		{
			name:             "Medium_Message",
			msgType:          "benchmark",
			payload:          map[string]int{"value": 123},
			expectedOverhead: 15 + len("benchmark"),
		},
		{
			name:             "Long_Message_Type",
			msgType:          "very_long_message_type_name_for_testing_protocol_overhead",
			payload:          "data",
			expectedOverhead: 15 + len("very_long_message_type_name_for_testing_protocol_overhead"),
		},
		{
			name:             "Complex_Payload",
			msgType:          "complex",
			payload: map[string]interface{}{
				"id":    12345,
				"name":  "测试用户",
				"data":  []int{1, 2, 3, 4, 5},
				"meta":  map[string]string{"version": "v0.0.2"},
			},
			expectedOverhead: 15 + len("complex"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 编码消息
			encoded, err := encodeMessage(c, tt.msgType, tt.payload, 12345)
			assert.NoError(t, err)

			// 计算负载大小
			payloadData, err := s.Serialize(tt.payload)
			assert.NoError(t, err)

			totalSize := len(encoded)
			payloadSize := len(payloadData)
			headerOverhead := totalSize - payloadSize
			overheadRatio := float64(headerOverhead) / float64(totalSize) * 100

			// 验证头部开销
			assert.Equal(t, tt.expectedOverhead, headerOverhead, 
				"头部开销不匹配: 消息类型=%s", tt.msgType)

			// 记录详细信息
			t.Logf("协议开销分析 - %s:", tt.name)
			t.Logf("  消息类型: %s (%d字节)", tt.msgType, len(tt.msgType))
			t.Logf("  总大小: %d字节", totalSize)
			t.Logf("  负载大小: %d字节", payloadSize)
			t.Logf("  头部开销: %d字节", headerOverhead)
			t.Logf("  开销比例: %.1f%%", overheadRatio)
			t.Logf("  协议版本: v0.0.2 (优化协议)")
		})
	}
}

// TestProtocolOverhead_Comparison 对比不同消息大小的协议开销
func TestProtocolOverhead_Comparison(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	// 定义不同大小的测试数据
	testCases := []struct {
		name        string
		payloadSize int
		description string
	}{
		{"Tiny", 10, "极小消息 (10字节)"},
		{"Small", 100, "小消息 (100字节)"},
		{"Medium", 1024, "中等消息 (1KB)"},
		{"Large", 10240, "大消息 (10KB)"},
		{"ExtraLarge", 102400, "超大消息 (100KB)"},
	}

	msgType := "size_test"
	expectedHeaderSize := 15 + len(msgType)

	t.Logf("协议开销对比分析 (v0.0.2):")
	t.Logf("固定头部开销: %d字节 (15字节基础 + %d字节消息类型)", 
		expectedHeaderSize, len(msgType))
	t.Logf("%-12s | %-8s | %-8s | %-8s | %-10s", 
		"消息大小", "总大小", "负载", "头部", "开销比例")
	t.Logf("%s", "-------------|----------|----------|----------|----------")

	for _, tc := range testCases {
		// 创建指定大小的测试数据
		testData := make([]byte, tc.payloadSize)
		for i := range testData {
			testData[i] = byte(i % 256)
		}

		payload := map[string]interface{}{
			"data": testData,
			"size": tc.payloadSize,
		}

		// 编码消息
		encoded, err := encodeMessage(c, msgType, payload, 12345)
		assert.NoError(t, err)

		// 计算开销
		payloadData, err := s.Serialize(payload)
		assert.NoError(t, err)

		totalSize := len(encoded)
		actualPayloadSize := len(payloadData)
		headerOverhead := totalSize - actualPayloadSize
		overheadRatio := float64(headerOverhead) / float64(totalSize) * 100

		// 验证头部开销
		assert.Equal(t, expectedHeaderSize, headerOverhead,
			"头部开销应该固定为 %d字节", expectedHeaderSize)

		// 输出对比数据
		t.Logf("%-12s | %-8d | %-8d | %-8d | %-9.1f%%", 
			tc.description, totalSize, actualPayloadSize, headerOverhead, overheadRatio)
	}
}

// TestProtocolOverhead_MessageTypeLength 测试不同消息类型长度的开销
func TestProtocolOverhead_MessageTypeLength(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	// 测试不同长度的消息类型
	messageTypes := []struct {
		name    string
		msgType string
		length  int
	}{
		{"Short", "a", 1},
		{"Normal", "test_message", 12},
		{"Long", "very_long_message_type_name", 27},
		{"Max", string(make([]byte, 255)), 255}, // 最大长度
	}

	payload := "test"
	expectedBaseOverhead := 15 // 固定头部大小

	t.Logf("消息类型长度对协议开销的影响:")
	t.Logf("%-8s | %-8s | %-8s | %-12s", 
		"类型长度", "总开销", "基础开销", "类型开销")
	t.Logf("%s", "---------|----------|----------|------------")

	for _, mt := range messageTypes {
		// 填充最大长度的消息类型
		if mt.length == 255 {
			for i := 0; i < 255; i++ {
				mt.msgType = mt.msgType[:i] + "x"
			}
		}

		encoded, err := encodeMessage(c, mt.msgType, payload, 12345)
		assert.NoError(t, err)

		payloadData, err := s.Serialize(payload)
		assert.NoError(t, err)

		totalOverhead := len(encoded) - len(payloadData)
		typeOverhead := totalOverhead - expectedBaseOverhead

		// 验证消息类型开销
		assert.Equal(t, len(mt.msgType), typeOverhead,
			"消息类型开销应该等于消息类型字节长度")

		t.Logf("%-8d | %-8d | %-8d | %-12d", 
			len(mt.msgType), totalOverhead, expectedBaseOverhead, typeOverhead)
	}
}

// TestProtocolOverhead_FlagImpact 测试标志位对开销的影响
func TestProtocolOverhead_FlagImpact(t *testing.T) {
	c := codec.NewLengthPrefixCodec(serializer.DefaultSerializer)

	msgType := "flag_test"
	payload := "test data"
	requestID := uint64(12345)

	// 测试不同标志组合
	flagTests := []struct {
		name  string
		flags uint8
	}{
		{"No_Flags", codec.FlagNone},
		{"Compressed", codec.FlagCompressed},
		{"Encrypted", codec.FlagEncrypted},
		{"Both", codec.FlagCompressed | codec.FlagEncrypted},
	}

	t.Logf("标志位对协议开销的影响:")
	for _, ft := range flagTests {
		encoded, err := encodeMessageWithFlags(c, msgType, payload, requestID, ft.flags)
		assert.NoError(t, err)

		payloadData, err := serializer.DefaultSerializer.Serialize(payload)
		assert.NoError(t, err)

		totalSize := len(encoded)
		payloadSize := len(payloadData)
		overhead := totalSize - payloadSize
		expectedOverhead := 15 + len(msgType) // 固定开销

		// 标志位不影响消息大小，只影响处理方式
		assert.Equal(t, expectedOverhead, overhead,
			"标志位不应影响协议开销")

		t.Logf("  %s (0x%02x): 总大小=%d字节, 开销=%d字节", 
			ft.name, ft.flags, totalSize, overhead)
	}
}

// 辅助函数：编码消息
func encodeMessage(c *codec.LengthPrefixCodec, msgType string, payload interface{}, requestID uint64) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := c.Encode(buf, msgType, payload, requestID)
	return buf.Bytes(), err
}

// 辅助函数：编码带标志位的消息
func encodeMessageWithFlags(c *codec.LengthPrefixCodec, msgType string, payload interface{}, requestID uint64, flags uint8) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := c.EncodeWithFlags(buf, msgType, payload, requestID, flags)
	return buf.Bytes(), err
}

// TestProtocolVersion_Validation 测试协议版本验证
func TestProtocolVersion_Validation(t *testing.T) {
	c := codec.NewLengthPrefixCodec(serializer.DefaultSerializer)

	// 编码一个正常消息
	encoded, err := encodeMessage(c, "version_test", "data", 12345)
	assert.NoError(t, err)
	assert.True(t, len(encoded) > 0)

	// 验证版本字段
	version := encoded[0]
	assert.Equal(t, uint8(codec.ProtocolVersion), version,
		"协议版本应该为 %d", codec.ProtocolVersion)

	t.Logf("协议版本验证:")
	t.Logf("  当前版本: v0.0.2")
	t.Logf("  协议版本号: %d", version)
	t.Logf("  版本字段位置: 字节0")
}