package coherency_cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCoherencyStorage(t *testing.T) {
	t.Run("Basic Operations", TestBasicOperations)
	t.Run("Concurrent Operations", TestConcurrentOperations)
	t.Run("Timeout Operations", TestTimeoutOperations)
	t.Run("Lock Operations", TestLockOperations)
	t.Run("Cache Miss Scenario", TestCacheMissScenario)
	t.Run("Nested CoherencyStorage", TestNestedCoherencyStorage)
	t.Run("Stress Test", TestStressTest)
}

func TestBasicOperations(t *testing.T) {
	cache := NewCoherencyStorage(
		NewMemoryStorage[string, string](),
		NewMemoryStorage[string, string](),
	)

	// Set Get Del Get
	key := "key"
	value := "value"

	// Test Set
	err := cache.Set(context.Background(), &key, &value, time.Millisecond*150)
	if err != nil {
		t.Fatal("Set failed:", err)
	}

	// Test Get
	getValue, err := cache.Get(context.Background(), &key, time.Second*10)
	if err != nil {
		t.Fatal("Get failed:", err)
	}

	if *getValue != value {
		t.Fatal("value not equal, expected:", value, "got:", *getValue)
	}

	// Test Del
	err = cache.Del(context.Background(), &key, time.Second*10)
	if err != nil {
		t.Fatal("Del failed:", err)
	}

	// Test Get after Del
	getValue, err = cache.Get(context.Background(), &key, time.Second*10)
	if err != nil {
		t.Fatal("Get after Del failed:", err)
	}
	if getValue != nil {
		t.Fatal("value should be nil after deletion, got:", *getValue)
	}
}

func TestConcurrentOperations(t *testing.T) {
	cache := NewCoherencyStorage(
		NewMemoryStorage[string, int](),
		NewMemoryStorage[string, int](),
	)

	const numGoroutines = 10
	const numOperations = 100
	var wg sync.WaitGroup

	// Option 1: Each goroutine uses different keys to avoid race conditions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "concurrent_key_" + string(rune(id)) // Each goroutine uses different key
				value := id*numOperations + j

				err := cache.Set(context.Background(), &key, &value, time.Millisecond*100)
				if err != nil {
					t.Errorf("Concurrent Set failed: %v", err)
					return
				}

				// Read verification
				getValue, err := cache.Get(context.Background(), &key, time.Millisecond*100)
				if err != nil {
					t.Errorf("Concurrent Get failed: %v", err)
					return
				}

				if getValue != nil && *getValue != value {
					t.Errorf("Value mismatch for key %s, expected: %d, got: %d", key, value, *getValue)
				}
			}
		}(i)
	}

	wg.Wait()

	// Option 2: Test concurrent writes to the same key (verify eventual consistency)
	t.Run("Same Key Concurrent Write", func(t *testing.T) {
		sameKey := "same_key"
		const numWriters = 5
		var wg2 sync.WaitGroup

		for i := 0; i < numWriters; i++ {
			wg2.Add(1)
			go func(writerID int) {
				defer wg2.Done()
				value := writerID * 100
				err := cache.Set(context.Background(), &sameKey, &value, time.Millisecond*100)
				if err != nil {
					t.Errorf("Same key concurrent Set failed: %v", err)
				}
			}(i)
		}

		wg2.Wait()

		// Verify eventual consistency: should get one of the written values
		finalValue, err := cache.Get(context.Background(), &sameKey, time.Millisecond*100)
		if err != nil {
			t.Fatal("Get final value failed:", err)
		}

		if finalValue == nil {
			t.Fatal("Expected final value, got nil")
		}

		// Verify final value is one of the written values (0, 100, 200, 300, 400)
		expectedValues := map[int]bool{0: true, 100: true, 200: true, 300: true, 400: true}
		if !expectedValues[*finalValue] {
			t.Errorf("Final value %d is not one of the expected values", *finalValue)
		}
	})
}

func TestTimeoutOperations(t *testing.T) {
	cache := NewCoherencyStorage(
		NewMemoryStorage[string, string](),
		NewMemoryStorage[string, string](),
	)

	key := "timeout_key"
	value := "timeout_value"

	// Test normal timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	err := cache.Set(ctx, &key, &value, time.Millisecond*100)
	if err != nil {
		t.Fatal("Set with timeout failed:", err)
	}

	// Test timeout scenario
	ctx, cancel = context.WithTimeout(context.Background(), time.Nanosecond*1)
	defer cancel()

	_, err = cache.Get(ctx, &key, time.Millisecond*100)
	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}
}

func TestLockOperations(t *testing.T) {
	cache := NewCoherencyStorage(
		NewMemoryStorage[string, string](),
		NewMemoryStorage[string, string](),
	)

	key := "lock_key"

	// Test basic lock functionality
	cache.Lock(context.Background(), &key)

	// Try to acquire lock in another goroutine
	lockAcquired := make(chan bool, 1)
	go func() {
		cache.Lock(context.Background(), &key)
		lockAcquired <- true
		cache.Unlock(context.Background(), &key)
	}()

	// Wait a short time to confirm lock is not acquired
	select {
	case <-lockAcquired:
		t.Fatal("Lock should not be acquired while another goroutine holds it")
	case <-time.After(time.Millisecond * 10):
		// Normal case, lock is not acquired
	}

	// Release lock
	cache.Unlock(context.Background(), &key)

	// Now lock should be acquirable
	select {
	case <-lockAcquired:
		// Normal case
	case <-time.After(time.Millisecond * 100):
		t.Fatal("Lock should be acquired after unlock")
	}
}

func TestCacheMissScenario(t *testing.T) {
	cache := NewCoherencyStorage(
		NewMemoryStorage[string, string](),
		NewMemoryStorage[string, string](),
	)

	key := "cache_miss_key"
	value := "cache_miss_value"

	// Set value only in source, not in cache
	err := cache.source.Set(context.Background(), &key, &value, time.Millisecond*100)
	if err != nil {
		t.Fatal("Source Set failed:", err)
	}

	// Get from cache (should miss, then get from source)
	getValue, err := cache.Get(context.Background(), &key, time.Second*10)
	if err != nil {
		t.Fatal("Get with cache miss failed:", err)
	}

	if getValue == nil {
		t.Fatal("Expected value from source, got nil")
	}

	if *getValue != value {
		t.Fatal("Value mismatch, expected:", value, "got:", *getValue)
	}

	// Get again should get from cache
	getValue2, err := cache.Get(context.Background(), &key, time.Second*10)
	if err != nil {
		t.Fatal("Second Get failed:", err)
	}

	if getValue2 == nil {
		t.Fatal("Expected value from cache, got nil")
	}

	if *getValue2 != value {
		t.Fatal("Value mismatch on second get, expected:", value, "got:", *getValue2)
	}
}

func TestNestedCoherencyStorage(t *testing.T) {
	// Create three-level nested CoherencyStorage
	// L1: Memory Cache -> L2: Memory Cache -> L3: Memory Storage
	L3Storage := NewMemoryStorage[string, string]()
	L2Cache := NewCoherencyStorage(
		NewMemoryStorage[string, string](),
		L3Storage,
	)
	L1Cache := NewCoherencyStorage(
		NewMemoryStorage[string, string](),
		L2Cache,
	)

	key := "nested_key"
	value := "nested_value"

	// Test 1: Set value in L3, then get from L1 (test multi-level cache penetration)
	t.Run("Multi-level Cache Penetration", func(t *testing.T) {
		// Set value in L3
		err := L3Storage.Set(context.Background(), &key, &value, time.Millisecond*100)
		if err != nil {
			t.Fatal("L3 Set failed:", err)
		}

		// Get from L1, should penetrate L1 and L2, get from L3
		getValue, err := L1Cache.Get(context.Background(), &key, time.Second*10)
		if err != nil {
			t.Fatal("L1 Get failed:", err)
		}

		if getValue == nil {
			t.Fatal("Expected value from L1, got nil")
		}

		if *getValue != value {
			t.Fatal("Value mismatch in L1, expected:", value, "got:", *getValue)
		}

		// Get from L1 again, should get from L1 cache
		getValue2, err := L1Cache.Get(context.Background(), &key, time.Second*10)
		if err != nil {
			t.Fatal("L1 second Get failed:", err)
		}

		if getValue2 == nil {
			t.Fatal("Expected value from L1 cache, got nil")
		}

		if *getValue2 != value {
			t.Fatal("Value mismatch in L1 cache, expected:", value, "got:", *getValue2)
		}
	})

	// Test 2: Set value in L1, verify consistency
	t.Run("L1 Set and Consistency", func(t *testing.T) {
		newKey := "l1_set_key"
		newValue := "l1_set_value"

		// Set value in L1
		err := L1Cache.Set(context.Background(), &newKey, &newValue, time.Millisecond*100)
		if err != nil {
			t.Fatal("L1 Set failed:", err)
		}

		// Get from L1
		getValue, err := L1Cache.Get(context.Background(), &newKey, time.Second*10)
		if err != nil {
			t.Fatal("L1 Get after Set failed:", err)
		}

		if getValue == nil {
			t.Fatal("Expected value from L1 after Set, got nil")
		}

		if *getValue != newValue {
			t.Fatal("Value mismatch in L1 after Set, expected:", newValue, "got:", *getValue)
		}

		// Get from L2 (should penetrate to L2)
		getValue2, err := L2Cache.Get(context.Background(), &newKey, time.Second*10)
		if err != nil {
			t.Fatal("L2 Get after L1 Set failed:", err)
		}

		if getValue2 == nil {
			t.Fatal("Expected value from L2 after L1 Set, got nil")
		}

		if *getValue2 != newValue {
			t.Fatal("Value mismatch in L2 after L1 Set, expected:", newValue, "got:", *getValue2)
		}

		// Get from L3 (should penetrate to L3)
		getValue3, err := L3Storage.Get(context.Background(), &newKey, time.Second*10)
		if err != nil {
			t.Fatal("L3 Get after L1 Set failed:", err)
		}

		if getValue3 == nil {
			t.Fatal("Expected value from L3 after L1 Set, got nil")
		}

		if *getValue3 != newValue {
			t.Fatal("Value mismatch in L3 after L1 Set, expected:", newValue, "got:", *getValue3)
		}
	})

	// Test 3: Concurrent operations on nested cache
	t.Run("Concurrent Nested Operations", func(t *testing.T) {
		const numGoroutines = 5
		const numOperations = 50
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					key := "concurrent_nested_key_" + string(rune(id))
					value := "value_" + string(rune(id)) + "_" + string(rune(j))

					// Set in L1
					err := L1Cache.Set(context.Background(), &key, &value, time.Millisecond*100)
					if err != nil {
						t.Errorf("Concurrent L1 Set failed: %v", err)
						return
					}

					// Get from L1 for verification
					getValue, err := L1Cache.Get(context.Background(), &key, time.Millisecond*100)
					if err != nil {
						t.Errorf("Concurrent L1 Get failed: %v", err)
						return
					}

					if getValue != nil && *getValue != value {
						t.Errorf("L1 Value mismatch for key %s, expected: %s, got: %s", key, value, *getValue)
					}

					// Get from L2 for verification
					getValue2, err := L2Cache.Get(context.Background(), &key, time.Millisecond*100)
					if err != nil {
						t.Errorf("Concurrent L2 Get failed: %v", err)
						return
					}

					if getValue2 != nil && *getValue2 != value {
						t.Errorf("L2 Value mismatch for key %s, expected: %s, got: %s", key, value, *getValue2)
					}
				}
			}(i)
		}

		wg.Wait()
	})

	// Test 4: Delete operation consistency
	t.Run("Delete Consistency", func(t *testing.T) {
		deleteKey := "delete_key"
		deleteValue := "delete_value"

		// Set value in L1
		err := L1Cache.Set(context.Background(), &deleteKey, &deleteValue, time.Millisecond*100)
		if err != nil {
			t.Fatal("L1 Set for delete test failed:", err)
		}

		// Verify value is set
		getValue, err := L1Cache.Get(context.Background(), &deleteKey, time.Second*10)
		if err != nil || getValue == nil || *getValue != deleteValue {
			t.Fatal("Value not set correctly for delete test")
		}

		// Delete in L1
		err = L1Cache.Del(context.Background(), &deleteKey, time.Millisecond*100)
		if err != nil {
			t.Fatal("L1 Del failed:", err)
		}

		// Verify deleted from L1
		getValue, err = L1Cache.Get(context.Background(), &deleteKey, time.Second*10)
		if err != nil || getValue != nil {
			t.Fatal("Value should be deleted from L1")
		}

		// Verify deleted from L2
		getValue2, err := L2Cache.Get(context.Background(), &deleteKey, time.Second*10)
		if err != nil || getValue2 != nil {
			t.Fatal("Value should be deleted from L2")
		}

		// Verify deleted from L3
		getValue3, err := L3Storage.Get(context.Background(), &deleteKey, time.Second*10)
		if err != nil || getValue3 != nil {
			t.Fatal("Value should be deleted from L3")
		}
	})
}

func TestStressTest(t *testing.T) {
	cache := NewCoherencyStorage(
		NewMemoryStorage[string, int](),
		NewMemoryStorage[string, int](),
	)

	const numKeys = 100
	const numOperations = 1000
	var wg sync.WaitGroup

	// Stress test: multiple goroutines operating on different keys simultaneously
	for i := 0; i < numKeys; i++ {
		wg.Add(1)
		go func(keyID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				key := "stress_key_" + string(rune(keyID))
				value := keyID*numOperations + j

				// Random operations: Set, Get, Del
				operation := j % 3
				switch operation {
				case 0: // Set
					err := cache.Set(context.Background(), &key, &value, time.Millisecond*50)
					if err != nil {
						t.Errorf("Stress Set failed for key %s: %v", key, err)
						return
					}
				case 1: // Get
					_, err := cache.Get(context.Background(), &key, time.Millisecond*50)
					if err != nil {
						t.Errorf("Stress Get failed for key %s: %v", key, err)
						return
					}
				case 2: // Del
					err := cache.Del(context.Background(), &key, time.Millisecond*50)
					if err != nil {
						t.Errorf("Stress Del failed for key %s: %v", key, err)
						return
					}
				}
			}
		}(i)
	}

	wg.Wait()
}

// Test MemoryStorage's independent functionality
func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage[string, int]()

	key := "test_key"
	value := 42

	// Test Set and Get
	err := storage.Set(context.Background(), &key, &value, time.Millisecond*100)
	if err != nil {
		t.Fatal("MemoryStorage Set failed:", err)
	}

	getValue, err := storage.Get(context.Background(), &key, time.Millisecond*100)
	if err != nil {
		t.Fatal("MemoryStorage Get failed:", err)
	}

	if getValue == nil {
		t.Fatal("Expected value, got nil")
	}

	if *getValue != value {
		t.Fatal("Value mismatch, expected:", value, "got:", *getValue)
	}

	// Test Del
	err = storage.Del(context.Background(), &key, time.Millisecond*100)
	if err != nil {
		t.Fatal("MemoryStorage Del failed:", err)
	}

	getValue, err = storage.Get(context.Background(), &key, time.Millisecond*100)
	if err != nil {
		t.Fatal("MemoryStorage Get after Del failed:", err)
	}

	if getValue != nil {
		t.Fatal("Expected nil after deletion, got:", *getValue)
	}

	// Test Timeout
	timeout := storage.Timeout()
	if timeout != time.Millisecond*100 {
		t.Fatal("Expected timeout 100ms, got:", timeout)
	}
}
