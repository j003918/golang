// hashset_test
package hashset

import (
	"testing"
)

func Test_Set(t *testing.T) {
	tmp := New()
	tmp.Set(1)
	tmp.Set("1")

	if tmp.Contains("1") && tmp.Contains("1") {
		t.Log("PASS")
	} else {
		t.Fatal("Fatal")
	}
}

func Test_Del(t *testing.T) {
	tmp := New()
	tmp.Set(1)
	tmp.Set("1")

	tmp.Del(1)
	tmp.Del("1")

	if tmp.Contains("1") && tmp.Contains("1") {
		t.Fatal("Fatal")
	} else {
		t.Log("PASS")
	}
}

func BenchmarkHashSet_Set(b *testing.B) {
	B := New()
	//B.Set(10)
	//B.Set(11)
	for i := 0; i < b.N; i++ {
		B.Set(i)
	}
}

func BenchmarkMap_Set(b *testing.B) {
	var B = make(map[interface{}]int, 3)
	//B[10] = 1
	//B[11] = 1
	for i := 0; i < b.N; i++ {
		B[i] = i
	}
}

func BenchmarkHashSet_Contains(b *testing.B) {
	B := New()
	B.Set(10)
	B.Set(11)
	for i := 0; i < b.N; i++ {
		if B.Contains(1) {

		}
		if B.Contains(11) {

		}
		if B.Contains(1000000) {

		}
	}
}

func BenchmarkMap_Contains(b *testing.B) {
	var B = make(map[interface{}]int, 3)
	B[10] = 1
	B[11] = 1
	for i := 0; i < b.N; i++ {
		if _, exists := B[1]; exists {

		}
		if _, exists := B[11]; exists {

		}
		if _, exists := B[1000000]; exists {

		}
	}
}
