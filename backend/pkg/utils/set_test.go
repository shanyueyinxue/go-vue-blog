package utils

import (
	"fmt"
	"sync"
	"testing"
)

// Set tests

func TestNewSet(t *testing.T) {
	s := NewSet(1, 2, 3)
	if s.Len() != 3 {
		t.Errorf("expected length 3, got %d", s.Len())
	}
	if !s.Contains(1) || !s.Contains(2) || !s.Contains(3) {
		t.Error("expected set to contain 1, 2, 3")
	}
}

func TestSetAdd(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Add(1) // duplicate
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
}

func TestSetRemove(t *testing.T) {
	s := NewSet(1, 2, 3)
	s.Remove(2)
	if s.Contains(2) {
		t.Error("expected 2 to be removed")
	}
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
}

func TestSetContains(t *testing.T) {
	s := NewSet("a", "b")
	if !s.Contains("a") {
		t.Error("expected set to contain 'a'")
	}
	if s.Contains("c") {
		t.Error("expected set to not contain 'c'")
	}
}

func TestSetClear(t *testing.T) {
	s := NewSet(1, 2, 3)
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("expected length 0 after clear, got %d", s.Len())
	}
}

func TestSetValues(t *testing.T) {
	s := NewSet(1, 2, 3)
	values := s.Values()
	if len(values) != 3 {
		t.Errorf("expected 3 values, got %d", len(values))
	}
}

func TestSetString(t *testing.T) {
	s := NewSet("c", "a", "b")
	str := s.String()
	expected := "{a; b; c}"
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}
}

func TestSetToString(t *testing.T) {
	s := NewSet(3, 1, 2)
	str := s.ToString()
	expected := "{1; 2; 3}"
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}
}

func TestSetConcurrency(t *testing.T) {
	s := NewSet[int]()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Add(i)
			s.Contains(i)
			s.Len()
			s.Values()
		}(i)
	}
	wg.Wait()
	if s.Len() != 100 {
		t.Errorf("expected length 100, got %d", s.Len())
	}
}

// HashSet tests

type testHashable struct {
	ID   string
	Name string
}

func (t testHashable) Hash() (string, error) {
	return t.ID, nil
}

type testToStringer struct {
	Value string
}

func (t testToStringer) ToString() string {
	return t.Value
}

func TestNewHashSet(t *testing.T) {
	s := NewHashSet(1, 2, 3)
	if s.Len() != 3 {
		t.Errorf("expected length 3, got %d", s.Len())
	}
}

func TestHashSetAdd(t *testing.T) {
	s := NewHashSet[int]()
	_ = s.Add(1)
	_ = s.Add(2)
	_ = s.Add(1) // duplicate
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
}

func TestHashSetRemove(t *testing.T) {
	s := NewHashSet("a", "b", "c")
	s.Remove("b")
	if s.Contains("b") {
		t.Error("expected 'b' to be removed")
	}
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
}

func TestHashSetContains(t *testing.T) {
	s := NewHashSet("x", "y")
	if !s.Contains("x") {
		t.Error("expected set to contain 'x'")
	}
	if s.Contains("z") {
		t.Error("expected set to not contain 'z'")
	}
}

func TestHashSetClear(t *testing.T) {
	s := NewHashSet(1, 2, 3)
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("expected length 0 after clear, got %d", s.Len())
	}
}

func TestHashSetValues(t *testing.T) {
	s := NewHashSet(1, 2, 3)
	values := s.Values()
	if len(values) != 3 {
		t.Errorf("expected 3 values, got %d", len(values))
	}
}

func TestHashSetString(t *testing.T) {
	s := NewHashSet("c", "a", "b")
	str := s.String()
	expected := "{a; b; c}"
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}
}

func TestHashSetToString(t *testing.T) {
	s := NewHashSet(3, 1, 2)
	str := s.ToString()
	expected := "{1; 2; 3}"
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}
}

func TestHashSetBasicTypeHash(t *testing.T) {
	s := NewHashSet[string]()
	_ = s.Add("hello")
	_ = s.Add("world")
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
	if !s.Contains("hello") {
		t.Error("expected set to contain 'hello'")
	}
}

func TestHashSetHashable(t *testing.T) {
	s := NewHashSet[testHashable]()
	_ = s.Add(testHashable{ID: "1", Name: "Alice"})
	_ = s.Add(testHashable{ID: "2", Name: "Bob"})
	_ = s.Add(testHashable{ID: "1", Name: "Alice2"}) // same ID, should overwrite
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
}

func TestHashSetToStringer(t *testing.T) {
	s := NewHashSet[testToStringer]()
	_ = s.Add(testToStringer{Value: "v1"})
	_ = s.Add(testToStringer{Value: "v2"})
	_ = s.Add(testToStringer{Value: "v1"}) // duplicate
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
}

func TestHashSetStruct(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}
	s := NewHashSet[User]()
	_ = s.Add(User{ID: 1, Name: "Alice"})
	_ = s.Add(User{ID: 2, Name: "Bob"})
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
	if !s.Contains(User{ID: 1, Name: "Alice"}) {
		t.Error("expected set to contain Alice")
	}
}

func TestHashSetConcurrency(t *testing.T) {
	s := NewHashSet[int]()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = s.Add(i)
			s.Contains(i)
			s.Len()
			s.Values()
		}(i)
	}
	wg.Wait()
	if s.Len() != 100 {
		t.Errorf("expected length 100, got %d", s.Len())
	}
}

func TestHashSetIntTypes(t *testing.T) {
	s8 := NewHashSet[int8]()
	_ = s8.Add(1)
	if !s8.Contains(1) {
		t.Error("expected int8 set to contain 1")
	}

	s64 := NewHashSet[int64]()
	_ = s64.Add(100)
	if !s64.Contains(100) {
		t.Error("expected int64 set to contain 100")
	}

	su := NewHashSet[uint]()
	_ = su.Add(42)
	if !su.Contains(42) {
		t.Error("expected uint set to contain 42")
	}
}

func TestHashSetFloatTypes(t *testing.T) {
	s32 := NewHashSet[float32]()
	_ = s32.Add(3.14)
	if !s32.Contains(3.14) {
		t.Error("expected float32 set to contain 3.14")
	}

	s64 := NewHashSet[float64]()
	_ = s64.Add(2.718)
	if !s64.Contains(2.718) {
		t.Error("expected float64 set to contain 2.718")
	}
}

func TestHashSetBoolType(t *testing.T) {
	s := NewHashSet[bool]()
	_ = s.Add(true)
	_ = s.Add(false)
	if s.Len() != 2 {
		t.Errorf("expected length 2, got %d", s.Len())
	}
	if !s.Contains(true) || !s.Contains(false) {
		t.Error("expected bool set to contain true and false")
	}
}

func TestHashSetStringSorted(t *testing.T) {
	s := NewHashSet("zebra", "apple", "mango")
	str := s.String()
	expected := "{apple; mango; zebra}"
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}
}

func ExampleNewSet() {
	s := NewSet(1, 2, 3)
	s.Add(4)
	fmt.Println(s.Contains(3))
	fmt.Println(s.Len())
	// Output:
	// true
	// 4
}

func ExampleNewHashSet() {
	s := NewHashSet("a", "b", "c")
	_ = s.Add("d")
	fmt.Println(s.Contains("a"))
	fmt.Println(s.Len())
	// Output:
	// true
	// 4
}
