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
	c := codec.NewBalancedCodec(s)

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
			expectedOverhead: 21, // Balanced协议固定头部: 4(Magic) + 1(Version+Flags+Length) + 8(RequestID) + 4(TypeID) + 4(Length) = 21字节
		},
		{
			name:             "Medium_Message",
			msgType:          "benchmark",
			payload:          map[string]int{"value": 123},
			expectedOverhead: 21, // Balanced协议固定头部
		},
		{
			name:             "Long_Message_Type",
			msgType:          "very_long_message_type_name_for_testing_protocol_overhead",
			payload:          "data",
			expectedOverhead: 21, // Balanced协议固定头部，TypeID固定4字节
		},
		{
			name:    "Complex_Payload",
			msgType: "complex",
			payload: map[string]interface{}{
				"id":   12345,
				"name": "测试用户",
				"data": []int{1, 2, 3, 4, 5},
				"meta": map[string]string{"version": "v0.0.2"},
			},
			expectedOverhead: 21, // Balanced协议固定头部
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
	c := codec.NewBalancedCodec(s)

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
	expectedHeaderSize := 21 // Balanced协议固定头部

	t.Logf("协议开销对比分析 (Balanced协议):")
	t.Logf("固定头部开销: %d字节 (4字节Magic + 1字节Version+Flags+Length + 8字节RequestID + 4字节TypeID + 4字节Length)",
		expectedHeaderSize)
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
	c := codec.NewBalancedCodec(s)

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
	expectedBaseOverhead := 21 // Balanced协议固定头部大小

	t.Logf("消息类型长度对协议开销的影响 (Balanced协议):")
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

		// Balanced协议中，TypeID固定为4字节，不随消息类型字符串长度变化
		assert.Equal(t, 0, typeOverhead,
			"Balanced协议中TypeID固定为4字节，不随消息类型字符串长度变化")

		t.Logf("%-8d | %-8d | %-8d | %-12d",
			len(mt.msgType), totalOverhead, expectedBaseOverhead, typeOverhead)
	}
}

// TestProtocolOverhead_FlagImpact 测试标志位对开销的影响
func TestProtocolOverhead_FlagImpact(t *testing.T) {
	// 创建带加密的编解码器
	key := []byte("1234567890123456") // 16字节密钥
	encryptor, err := codec.NewAESEncryptor(key)
	assert.NoError(t, err)

	c := codec.NewBalancedCodecWithEncryption(serializer.DefaultSerializer, encryptor)

	msgType := "flag_test"
	payload := "test data"
	requestID := uint64(12345)

	// 测试不同标志组合
	flagTests := []struct {
		name  string
		flags uint8
	}{
		{"No_Flags", codec.BalancedFlagNone},
		{"Compressed", codec.BalancedFlagCompressed},
		{"Encrypted", codec.BalancedFlagEncrypted},
		{"Both", codec.BalancedFlagCompressed | codec.BalancedFlagEncrypted},
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

		// 根据标志位计算期望开销
		var expectedOverhead int
		if ft.flags&codec.BalancedFlagEncrypted != 0 {
			// 加密会增加 AES-GCM 的 nonce(12字节) + tag(16字节) = 28字节
			expectedOverhead = 21 + 28 // 基础开销 + 加密开销
		} else {
			expectedOverhead = 21 // Balanced协议固定开销
		}

		// 验证开销
		assert.Equal(t, expectedOverhead, overhead,
			"标志位开销不匹配: %s", ft.name)

		t.Logf("  %s (0x%02x): 总大小=%d字节, 开销=%d字节",
			ft.name, ft.flags, totalSize, overhead)
	}
}

// 辅助函数：编码消息
func encodeMessage(c *codec.BalancedCodec, msgType string, payload interface{}, requestID uint64) ([]byte, error) {
	buf := &bytes.Buffer{}
	// 将字符串转换为哈希ID
	typeID := hashString(msgType)
	err := c.Encode(buf, typeID, payload, requestID)
	return buf.Bytes(), err
}

// 辅助函数：编码带标志位的消息
func encodeMessageWithFlags(c *codec.BalancedCodec, msgType string, payload interface{}, requestID uint64, flags uint8) ([]byte, error) {
	buf := &bytes.Buffer{}
	// 将字符串转换为哈希ID
	typeID := hashString(msgType)
	err := c.EncodeWithFlags(buf, typeID, payload, requestID, flags, nil)
	return buf.Bytes(), err
}

// TestProtocolVersion_Validation 测试协议版本验证
func TestProtocolVersion_Validation(t *testing.T) {
	c := codec.NewBalancedCodec(serializer.DefaultSerializer)

	// 编码一个正常消息
	encoded, err := encodeMessage(c, "version_test", "data", 12345)
	assert.NoError(t, err)
	assert.True(t, len(encoded) > 0)

	// 验证版本字段 (Balanced协议中版本在字节4的高4位)
	versionByte := encoded[4]
	version := (versionByte >> 4) & 0x0F
	assert.Equal(t, uint8(codec.BalancedVersion), version,
		"协议版本应该为 %d", codec.BalancedVersion)

	t.Logf("协议版本验证:")
	t.Logf("  当前版本: Balanced协议")
	t.Logf("  协议版本号: %d", version)
	t.Logf("  版本字段位置: 字节4的高4位")
}
