package core

import (
	"errors"
	"hash/fnv"
	"sync"
)

var ErrTypeConflict = errors.New("type hash conflict")

// Registry 类型注册器，将string转换为uint32 ID
type Registry struct {
	nameToID map[string]uint32 // string -> int ID (hash)
	idToName map[uint32]string // int ID -> string
	mutex    sync.RWMutex
}

// NewRegistry 创建新的类型注册器
func NewRegistry() *Registry {
	return &Registry{
		nameToID: make(map[string]uint32),
		idToName: make(map[uint32]string),
	}
}

// Register 将string转换为uint32 ID
func (r *Registry) Register(msgType string) (uint32, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 计算哈希
	h := fnv.New32a()
	if _, err := h.Write([]byte(msgType)); err != nil {
		return 0, err
	}

	hash := h.Sum32()

	// 检查冲突
	if existing, exists := r.idToName[hash]; exists && existing != msgType {
		return 0, ErrTypeConflict // 冲突错误
	}

	// 注册
	r.nameToID[msgType] = hash
	r.idToName[hash] = msgType
	return hash, nil
}

// GetName 根据ID获取string
func (r *Registry) GetName(id uint32) (string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	name, exists := r.idToName[id]
	return name, exists
}

// GetID 根据string获取ID
func (r *Registry) GetID(msgType string) (uint32, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	id, exists := r.nameToID[msgType]
	return id, exists
}

// GetAllTypes 获取所有注册的类型（用于调试）
func (r *Registry) GetAllTypes() map[string]uint32 {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 返回副本
	types := make(map[string]uint32)
	for k, v := range r.nameToID {
		types[k] = v
	}
	return types
}

// Clear 清空所有注册的类型（用于测试）
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.nameToID = make(map[string]uint32)
	r.idToName = make(map[uint32]string)
}
