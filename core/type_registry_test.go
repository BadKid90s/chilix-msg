package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeRegistry(t *testing.T) {
	registry := NewRegistry()

	// 测试类型注册
	msgType := "new_message_type"
	typeID, err := registry.Register(msgType)
	require.NoError(t, err)
	assert.NotZero(t, typeID)

	// 测试重复注册
	typeID2, err := registry.Register(msgType)
	require.NoError(t, err)
	assert.Equal(t, typeID, typeID2)

	// 测试ID到名称的转换
	name, exists := registry.GetName(typeID)
	require.True(t, exists)
	assert.Equal(t, msgType, name)

	// 测试名称到ID的转换
	id, exists := registry.GetID(msgType)
	require.True(t, exists)
	assert.Equal(t, typeID, id)
}

func TestTypeRegistryConflict(t *testing.T) {
	registry := NewRegistry()

	// 注册第一个类型
	msgType1 := "test_type"
	typeID1, err := registry.Register(msgType1)
	require.NoError(t, err)

	// 尝试注册冲突的类型（虽然FNV-1a冲突概率很低，但我们可以测试错误处理）
	// 这里我们测试一个不存在的冲突情况
	msgType2 := "different_type"
	typeID2, err := registry.Register(msgType2)
	require.NoError(t, err)
	assert.NotEqual(t, typeID1, typeID2)
}

func TestTypeRegistryGetAllTypes(t *testing.T) {
	registry := NewRegistry()

	// 注册多个类型
	types := []string{"type1", "type2", "type3"}
	expectedTypes := make(map[string]uint32)

	for _, msgType := range types {
		typeID, err := registry.Register(msgType)
		require.NoError(t, err)
		expectedTypes[msgType] = typeID
	}

	// 获取所有类型
	allTypes := registry.GetAllTypes()
	assert.Equal(t, len(expectedTypes), len(allTypes))

	for msgType, expectedID := range expectedTypes {
		actualID, exists := allTypes[msgType]
		assert.True(t, exists)
		assert.Equal(t, expectedID, actualID)
	}
}

func TestTypeRegistryClear(t *testing.T) {
	registry := NewRegistry()

	// 注册一些类型
	msgType := "test_type"
	typeID, err := registry.Register(msgType)
	require.NoError(t, err)
	assert.NotZero(t, typeID)

	// 验证类型存在
	_, exists := registry.GetID(msgType)
	assert.True(t, exists)

	// 清空注册器
	registry.Clear()

	// 验证类型不存在
	_, exists = registry.GetID(msgType)
	assert.False(t, exists)

	// 验证所有类型为空
	allTypes := registry.GetAllTypes()
	assert.Empty(t, allTypes)
}

func TestTypeRegistryConcurrent(t *testing.T) {
	registry := NewRegistry()

	// 并发注册多个类型
	types := []string{"type1", "type2", "type3", "type4", "type5"}
	done := make(chan bool, len(types))

	for _, msgType := range types {
		go func(mt string) {
			defer func() { done <- true }()

			typeID, err := registry.Register(mt)
			require.NoError(t, err)
			assert.NotZero(t, typeID)

			// 验证注册成功
			name, exists := registry.GetName(typeID)
			assert.True(t, exists)
			assert.Equal(t, mt, name)
		}(msgType)
	}

	// 等待所有goroutine完成
	for i := 0; i < len(types); i++ {
		<-done
	}

	// 验证所有类型都已注册
	allTypes := registry.GetAllTypes()
	assert.Equal(t, len(types), len(allTypes))
}
