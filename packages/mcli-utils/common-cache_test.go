package mcliutils

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCCache_TestParallel(t *testing.T) {
	var err error
	var wg sync.WaitGroup

	cc := NewCCache(10000, func(params ...interface{}) (interface{}, error) {
		if len(params) == 0 {
			return nil, fmt.Errorf("no params provided")
		}
		if len(params) == 1 {
			valToProcess := params[0].(string)
			return strings.ToUpper(valToProcess), nil
		}
		var result []string = make([]string, 0, len(params))
		for _, param := range params {
			valToProcess := param.(string)
			result = append(result, strings.ToUpper(valToProcess))
		}
		return result, nil
	})

	values := []interface{}{"val1.1", "val1.2"}
	_, err = cc.Set(testingkey, values...)
	if err != nil {
		t.Errorf("Expected no error for non-existent key1, but got error: %v", err)
	}
	_, err = cc.Set("key2", "val2")
	if err != nil {
		t.Errorf("Expected no error for non-existent key2, but got error: %v", err)
	}
	wg.Add(1)
	go func(key string) {
		defer wg.Done()
		key1Value, err := cc.GetAndSetIfNotExists(testingkey)
		t.Log("val from "+key+":", key1Value, err, "count: ", cc.Cache[key].Count)
	}(testingkey)

	wg.Add(1)
	go func(key string) {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			key2Value, err := cc.GetAndSetIfNotExists(key)
			t.Log("val from "+key+":", key2Value, err, "count: ", cc.Cache[key].Count)
			time.Sleep(300 * time.Millisecond)
		}
	}("key2")

	wg.Wait()
}

func TestCCache_GetAndSetIfNotExists(t *testing.T) {
	cc := NewCCache(1000, nil)

	// Test case 1: Get non-existent key
	_, err := cc.GetAndSetIfNotExists("key1")
	if err == nil {
		t.Error("Expected error for non-existent key, but got nil")
	}

	// Test case 2: Set a value
	val, err := cc.GetAndSetIfNotExists("key1", "value1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', but got %v", val)
	}

	// Test case 2: Set a int value
	val, err = cc.GetAndSetIfNotExists("key100", 100)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != 100 {
		t.Errorf("Expected 100, but got %v", val)
	}

	// Test case 3: Get an existing value
	val, err = cc.GetAndSetIfNotExists("key1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', but got %v", val)
	}

	// Test case 4: Set new value for an existing key
	val, err = cc.Set("key1", "newvalue1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != "newvalue1" {
		t.Errorf("Expected 'newvalue1', but got %v", val)
	}
}

func TestCCache_Remove(t *testing.T) {
	cc := NewCCache(1000, nil)

	// Test case 1: Remove non-existent key
	err := cc.Remove("key1")
	if err == nil {
		t.Error("Expected error for non-existent key, but got nil")
	}

	// Test case 2: Set a value and remove it
	_, err = cc.GetAndSetIfNotExists("key1", "value1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = cc.Remove("key1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test case 3: Verify key is removed
	_, err = cc.GetAndSetIfNotExists("key1")
	if err == nil {
		t.Error("Expected error for non-existent key, but got nil")
	}
}

func TestCCache_Optimize(t *testing.T) {
	cc := NewCCache(100, nil)
	cc.MaxEntries = 3
	// Set some values
	cc.GetAndSetIfNotExists("key1", "value1")
	cc.GetAndSetIfNotExists("key2", "value2")
	cc.GetAndSetIfNotExists("key3", "value3")
	cc.GetAndSetIfNotExists("key4", "value4")
	cc.GetAndSetIfNotExists("key5", "value5")
	cc.GetAndSetIfNotExists("key6", "value6")

	if _, err := cc.GetAndSetIfNotExists("key3"); err != nil {
		t.Errorf("Expected no error for expired key, but got not nil error: %v", err)
	}
	if _, err := cc.GetAndSetIfNotExists("key3"); err != nil {
		t.Errorf("Expected no error for expired key, but got not nil error: %v", err)
	}
	if _, err := cc.GetAndSetIfNotExists("key1"); err != nil {
		t.Errorf("Expected no error for expired key, but got not nil error: %v", err)
	}
	if _, err := cc.GetAndSetIfNotExists("key1"); err != nil {
		t.Errorf("Expected no error for expired key, but got not nil error: %v", err)
	}
	if _, err := cc.GetAndSetIfNotExists("key1"); err != nil {
		t.Errorf("Expected no error for expired key, but got not nil error: %v", err)
	}
	if _, err := cc.GetAndSetIfNotExists("key5"); err != nil {
		t.Errorf("Expected no error for expired key, but got not nil error: %v", err)
	}

	// Optimize cache
	_ = cc.Optimize()
	if len(cc.Cache) > 3 {
		t.Errorf("Expected cache optimization to be successful, and count of entries less then 3, got count equal to len(cc.Cache)")
	}

	// Sleep for a while to allow entries to expire
	time.Sleep(200 * time.Millisecond)

	_ = cc.Optimize()

	// Verify that expired entries are removed
	if _, err := cc.GetAndSetIfNotExists("key1"); err == nil {
		t.Error("Expected error for expired key, but got nil")
	}
	if _, err := cc.GetAndSetIfNotExists("key2"); err == nil {
		t.Error("Expected error for expired key, but got nil")
	}
}
