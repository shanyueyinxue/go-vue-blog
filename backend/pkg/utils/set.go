package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Set[T comparable] struct {
	mu       sync.RWMutex
	elements map[T]struct{}
}

func NewSet[T comparable](v ...T) *Set[T] {
	s := &Set[T]{
		elements: make(map[T]struct{}),
	}
	for _, val := range v {
		s.Add(val)
	}
	return s
}

func (s *Set[T]) Add(val T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements[val] = struct{}{}
}

func (s *Set[T]) Remove(val T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.elements, val)
}

func (s *Set[T]) Contains(val T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.elements[val]
	return ok
}

func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.elements)
}

func (s *Set[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements = make(map[T]struct{})
}

func (s *Set[T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]T, 0, len(s.elements))
	for key := range s.elements {
		values = append(values, key)
	}
	return values
}

func (s *Set[T]) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]string, 0, len(s.elements))
	for key := range s.elements {
		values = append(values, fmt.Sprintf("%v", key))
	}
	sort.Strings(values)
	return "{" + strings.Join(values, "; ") + "}"
}

func (s *Set[T]) ToString() string {
	return s.String()
}

// StructSet -----------------------------------------------------

type Hashable interface {
	Hash() (string, error)
}
type ToStringer interface {
	ToString() string
}

type HashSet[T any] struct {
	mu       sync.RWMutex
	elements map[string]T
}

// NewHashSet 创建一个 HashSet;
// 注意：T 可以是任何类型，可以选择实现 Hashable 接口 或 ToStringer 接口，没有实现的类型会使用 json 序列化后计算 hash 值
func NewHashSet[T any](v ...T) *HashSet[T] {
	s := &HashSet[T]{
		elements: make(map[string]T),
	}
	for _, val := range v {
		s.Add(val)
	}
	return s
}

func (s *HashSet[T]) isBasicType(value T) (string, bool) {
	switch v := any(value).(type) {
	case string:
		return v, true
	case int:
		return fmt.Sprintf("%d", v), true
	case int8:
		return fmt.Sprintf("%d", v), true
	case int16:
		return fmt.Sprintf("%d", v), true
	case int32:
		return fmt.Sprintf("%d", v), true
	case int64:
		return fmt.Sprintf("%d", v), true
	case uint:
		return fmt.Sprintf("%d", v), true
	case uint8:
		return fmt.Sprintf("%d", v), true
	case uint16:
		return fmt.Sprintf("%d", v), true
	case uint32:
		return fmt.Sprintf("%d", v), true
	case uint64:
		return fmt.Sprintf("%d", v), true
	case float32:
		return fmt.Sprintf("%v", v), true
	case float64:
		return fmt.Sprintf("%v", v), true
	case bool:
		return fmt.Sprintf("%t", v), true
	default:
		return "", false
	}
}

func (s *HashSet[T]) hashValue(value T) (string, error) {
	if key, ok := s.isBasicType(value); ok {
		return key, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (s *HashSet[T]) hashKey(value T) (string, error) {
	hashable, ok := any(value).(Hashable)
	if ok {
		return hashable.Hash()
	}
	toStr, ok := any(value).(ToStringer)
	if ok {
		return toStr.ToString(), nil
	}
	return s.hashValue(value)
}

func (s *HashSet[T]) Add(val T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.elements == nil {
		s.elements = make(map[string]T)
	}
	key, err := s.hashKey(val)
	if err != nil {
		return err
	}
	s.elements[key] = val
	return nil
}

func (s *HashSet[T]) Remove(val T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key, err := s.hashKey(val)
	if err != nil {
		return
	}
	if _, ok := s.elements[key]; !ok {
		return
	}
	delete(s.elements, key)
}

func (s *HashSet[T]) Contains(val T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key, err := s.hashKey(val)
	if err != nil {
		return false
	}
	_, ok := s.elements[key]
	return ok
}

func (s *HashSet[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.elements)
}

func (s *HashSet[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements = make(map[string]T)
}

func (s *HashSet[T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]T, 0, len(s.elements))
	for _, val := range s.elements {
		values = append(values, val)
	}
	return values
}

func (s *HashSet[T]) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]string, 0, len(s.elements))
	for _, val := range s.elements {
		values = append(values, fmt.Sprintf("%v", val))
	}
	sort.Strings(values)
	return "{" + strings.Join(values, "; ") + "}"
}

func (s *HashSet[T]) ToString() string {
	return s.String()
}
