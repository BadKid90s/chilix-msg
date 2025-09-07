package core

import (
	"errors"
	"hash/fnv"
	"sync"
)

var ErrTypeConflict = errors.New("type hash conflict")

// TypeRegistry 类型注册器，将string转换为uint32 ID
type TypeRegistry struct {
	nameToID map[string]uint32 // string -> int ID (hash)
	idToName map[uint32]string // int ID -> string
	mutex    sync.RWMutex
}

// NewTypeRegistry 创建新的类型注册器
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		nameToID: make(map[string]uint32),
		idToName: make(map[uint32]string),
	}
}

// Register 将string转换为uint32 ID
func (tr *TypeRegistry) Register(msgType string) (uint32, error) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	// 计算哈希
	h := fnv.New32a()
	if _, err := h.Write([]byte(msgType)); err != nil {
		return 0, err
	}

	hash := h.Sum32()

	// 检查冲突
	if existing, exists := tr.idToName[hash]; exists && existing != msgType {
		return 0, ErrTypeConflict // 冲突错误
	}

	// 注册
	tr.nameToID[msgType] = hash
	tr.idToName[hash] = msgType
	return hash, nil
}

// GetName 根据ID获取string
func (tr *TypeRegistry) GetName(id uint32) (string, bool) {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()
	name, exists := tr.idToName[id]
	return name, exists
}

// GetID 根据string获取ID
func (tr *TypeRegistry) GetID(msgType string) (uint32, bool) {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()
	id, exists := tr.nameToID[msgType]
	return id, exists
}

// GetAllTypes 获取所有注册的类型（用于调试）
func (tr *TypeRegistry) GetAllTypes() map[string]uint32 {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()

	// 返回副本
	types := make(map[string]uint32)
	for k, v := range tr.nameToID {
		types[k] = v
	}
	return types
}

// Clear 清空所有注册的类型（用于测试）
func (tr *TypeRegistry) Clear() {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	tr.nameToID = make(map[string]uint32)
	tr.idToName = make(map[uint32]string)
}
