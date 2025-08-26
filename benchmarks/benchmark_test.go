package benchmarks

import (
	"bytes"
	"testing"

	"github.com/BadKid90s/chilix-msg/codec"
	"github.com/BadKid90s/chilix-msg/serializer"
)

// 性能基准测试套件
// 用于监控协议性能变化，验证优化效果

// BenchmarkCodec_Encode_SmallMessage 测试小消息编码性能
func BenchmarkCodec_Encode_SmallMessage(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "test"
	payload := "hello"  // ~10字节消息
	requestID := uint64(12345)

	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = c.Encode(buf, msgType, payload, requestID)
	}
}

// BenchmarkCodec_Encode_MediumMessage 测试中等消息编码性能
func BenchmarkCodec_Encode_MediumMessage(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "benchmark_test"
	payload := map[string]interface{}{
		"id":      12345,
		"message": "这是一个性能测试消息",
		"data":    []int{1, 2, 3, 4, 5},
		"meta": map[string]string{
			"version": "v0.0.2",
			"type":    "test",
		},
	}
	requestID := uint64(9876543210)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = c.Encode(buf, msgType, payload, requestID)
	}
}

// BenchmarkCodec_Encode_LargeMessage 测试大消息编码性能
func BenchmarkCodec_Encode_LargeMessage(b *testing.B) {
	s := serializer.DefaultSerializer  
	c := codec.NewLengthPrefixCodec(s)

	msgType := "large_test"
	// 创建1KB的测试数据
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	
	payload := map[string]interface{}{
		"id":        12345,
		"timestamp": "2025-08-27T07:00:00Z",
		"data":      largeData,
		"checksum":  "abc123def456",
	}
	requestID := uint64(9876543210)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = c.Encode(buf, msgType, payload, requestID)
	}
}

// BenchmarkCodec_Encode_ExtraLargeMessage 测试超大消息编码性能
func BenchmarkCodec_Encode_ExtraLargeMessage(b *testing.B) {
	s := serializer.DefaultSerializer  
	c := codec.NewLengthPrefixCodec(s)

	msgType := "xl_test"
	// 创建10KB的测试数据
	largeData := make([]byte, 10*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	
	payload := map[string]interface{}{
		"id":        12345,
		"timestamp": "2025-08-27T07:00:00Z",
		"data":      largeData,
		"metadata": map[string]interface{}{
			"size":     len(largeData),
			"encoding": "binary",
			"version":  "v0.0.2",
		},
	}
	requestID := uint64(9876543210)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = c.Encode(buf, msgType, payload, requestID)
	}
}

// BenchmarkCodec_Decode_SmallMessage 测试小消息解码性能
func BenchmarkCodec_Decode_SmallMessage(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "test"
	payload := "hello"
	requestID := uint64(12345)

	// 预编码消息
	buf := &bytes.Buffer{}
	_ = c.Encode(buf, msgType, payload, requestID)
	encodedData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(encodedData)
		_, _, _, _ = c.Decode(reader)
	}
}

// BenchmarkCodec_Decode_MediumMessage 测试中等消息解码性能
func BenchmarkCodec_Decode_MediumMessage(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "benchmark_test"
	payload := map[string]interface{}{
		"id":      12345,
		"message": "这是一个性能测试消息",
		"data":    []int{1, 2, 3, 4, 5},
		"meta": map[string]string{
			"version": "v0.0.2",
			"type":    "test",
		},
	}
	requestID := uint64(9876543210)

	// 预编码消息
	buf := &bytes.Buffer{}
	_ = c.Encode(buf, msgType, payload, requestID)
	encodedData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(encodedData)
		_, _, _, _ = c.Decode(reader)
	}
}

// BenchmarkCodec_Decode_LargeMessage 测试大消息解码性能
func BenchmarkCodec_Decode_LargeMessage(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "large_test"
	// 创建1KB的测试数据
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	
	payload := map[string]interface{}{
		"id":        12345,
		"timestamp": "2025-08-27T07:00:00Z",
		"data":      largeData,
		"checksum":  "abc123def456",
	}
	requestID := uint64(9876543210)

	// 预编码消息
	buf := &bytes.Buffer{}
	_ = c.Encode(buf, msgType, payload, requestID)
	encodedData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(encodedData)
		_, _, _, _ = c.Decode(reader)
	}
}

// BenchmarkCodec_EncodeWithFlags 测试带标志位的编码性能
func BenchmarkCodec_EncodeWithFlags(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "test_with_flags"
	payload := map[string]interface{}{
		"message": "带标志位的测试消息",
		"flags":   []string{"compressed", "encrypted"},
	}
	requestID := uint64(12345)
	flags := uint8(codec.FlagCompressed | codec.FlagEncrypted)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		_ = c.EncodeWithFlags(buf, msgType, payload, requestID, flags)
	}
}

// BenchmarkCodec_DecodeWithFlags 测试带标志位的解码性能
func BenchmarkCodec_DecodeWithFlags(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "test_with_flags"
	payload := map[string]interface{}{
		"message": "带标志位的测试消息",
		"flags":   []string{"compressed", "encrypted"},
	}
	requestID := uint64(12345)
	flags := uint8(codec.FlagCompressed | codec.FlagEncrypted)

	// 预编码消息
	buf := &bytes.Buffer{}
	_ = c.EncodeWithFlags(buf, msgType, payload, requestID, flags)
	encodedData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(encodedData)
		_, _, _, _, _ = c.DecodeWithFlags(reader)
	}
}

// BenchmarkCodec_RoundTrip 测试完整的编解码往返性能
func BenchmarkCodec_RoundTrip(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "roundtrip_test"
	payload := map[string]interface{}{
		"id":      12345,
		"message": "往返测试消息",
		"data":    []int{1, 2, 3, 4, 5},
	}
	requestID := uint64(9876543210)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 编码
		buf := &bytes.Buffer{}
		_ = c.Encode(buf, msgType, payload, requestID)
		
		// 解码
		_, _, _, _ = c.Decode(buf)
	}
}

// BenchmarkCodec_ConcurrentEncode 测试并发编码性能
func BenchmarkCodec_ConcurrentEncode(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "concurrent_test"
	payload := map[string]string{"message": "并发测试"}
	requestID := uint64(12345)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := &bytes.Buffer{}
			_ = c.Encode(buf, msgType, payload, requestID)
		}
	})
}

// BenchmarkCodec_ConcurrentDecode 测试并发解码性能
func BenchmarkCodec_ConcurrentDecode(b *testing.B) {
	s := serializer.DefaultSerializer
	c := codec.NewLengthPrefixCodec(s)

	msgType := "concurrent_test"
	payload := map[string]string{"message": "并发测试"}
	requestID := uint64(12345)

	// 预编码消息
	buf := &bytes.Buffer{}
	_ = c.Encode(buf, msgType, payload, requestID)
	encodedData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reader := bytes.NewReader(encodedData)
			_, _, _, _ = c.Decode(reader)
		}
	})
}