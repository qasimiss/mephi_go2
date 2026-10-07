package main

import (
	"encoding/binary"
	"math"
	"slices"
	"testing"
)

func TestGenerateRandomElements(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantLen int
	}{
		{"нулевой размер", 0, 0},
		{"отрицательный размер", -10, 0},
		{"один элемент", 1, 1},
		{"обычный размер", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateRandomElements(tt.size)
			if len(got) != tt.wantLen {
				t.Fatalf("длина слайса = %d, ожидается %d", len(got), tt.wantLen)
			}
			for i, value := range got {
				if value < 0 {
					t.Fatalf("элемент[%d] = %d, ожидается неотрицательное число", i, value)
				}
			}
		})
	}
}

func TestGenerateRandomElementsRandomness(t *testing.T) {
	first := generateRandomElements(1000)
	second := generateRandomElements(1000)

	same := true
	for i := range first {
		if first[i] != second[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("два вызова generateRandomElements вернули одинаковые слайсы")
	}
}

func TestMaximum(t *testing.T) {
	tests := []struct {
		name string
		data []int
		want int
	}{
		{"пустой слайс", []int{}, 0},
		{"nil-слайс", nil, 0},
		{"один элемент", []int{42}, 42},
		{"максимум в начале", []int{100, 1, 2, 3}, 100},
		{"максимум в середине", []int{1, 100, 2}, 100},
		{"максимум в конце", []int{1, 2, 3, 100}, 100},
		{"все элементы одинаковые", []int{7, 7, 7}, 7},
		{"отрицательные числа", []int{-5, -1, -10}, -1},
		{"нуль среди элементов", []int{0, 0, 1}, 1},
		{"максимальное значение int", []int{1, math.MaxInt, 2}, math.MaxInt},
		{"минимальное значение int", []int{math.MinInt}, math.MinInt},
		{"MinInt среди элементов", []int{math.MinInt, -5}, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maximum(tt.data); got != tt.want {
				t.Fatalf("maximum(%v) = %d, ожидается %d", tt.data, got, tt.want)
			}
		})
	}
}

func TestMaxChunks(t *testing.T) {
	tests := []struct {
		name string
		data []int
		want int
	}{
		{"пустой слайс", []int{}, 0},
		{"nil-слайс", nil, 0},
		{"один элемент", []int{42}, 42},
		{"максимум в начале", []int{100, 1, 2, 3}, 100},
		{"максимум в середине", []int{1, 100, 2}, 100},
		{"максимум в конце", []int{1, 2, 3, 100}, 100},
		{"все элементы одинаковые", []int{7, 7, 7}, 7},
		{"отрицательные числа", []int{-5, -1, -10}, -1},
		{"максимальное значение int", []int{1, math.MaxInt, 2}, math.MaxInt},
		{"минимальное значение int", []int{math.MinInt}, math.MinInt},
		{"MinInt среди элементов", []int{math.MinInt, -5}, -5},
		{"длина 7 — меньше CHUNKS", []int{9, 5, 7, 1, 3, 8, 2}, 9},
		{"длина 8 — ровно CHUNKS", []int{1, 2, 3, 4, 5, 6, 7, 8}, 8},
		{"длина 9 — остаток 1", []int{9, 1, 1, 1, 1, 1, 1, 1, 1}, 9},
		{"длина 10 — остаток 2, максимум в потерянном хвосте", []int{1, 1, 1, 1, 1, 1, 1, 1, 2, 3}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxChunks(tt.data); got != tt.want {
				t.Fatalf("maxChunks(%v) = %d, ожидается %d", tt.data, got, tt.want)
			}
		})
	}
}

// TestMaxChunksEqualsMaximum — свойство согласованности: на случайных данных
// фиксированной длины результаты maximum и maxChunks обязаны совпадать.
func TestMaxChunksEqualsMaximum(t *testing.T) {
	for _, size := range []int{1, 2, 7, 8, 9, 10, 15, 16, 17, 100, 1001, 100_000} {
		data := generateRandomElements(size)
		want := maximum(data)
		if got := maxChunks(data); got != want {
			t.Fatalf("size=%d: maxChunks = %d, maximum = %d", size, got, want)
		}
	}
}

// TestMaxChunksConcurrent — многократный запуск под гонками (go test -race):
// ловит состояние гонки между горутинами и WaitGroup.
func TestMaxChunksConcurrent(t *testing.T) {
	data := generateRandomElements(100_000)
	want := maximum(data)

	for i := 0; i < 100; i++ {
		if got := maxChunks(data); got != want {
			t.Fatalf("итерация %d: maxChunks = %d, ожидается %d", i, got, want)
		}
	}
}

// bytesToInts детерминированно декодирует байты в слайс целых чисел:
// каждые 8 байт (включая неполный хвост) становятся одним элементом,
// поэтому из байтов получается слайс любой длины.
func bytesToInts(raw []byte) []int {
	const size = 8 // sizeof(int64)

	data := make([]int, 0, len(raw)/size+1)
	for i := 0; i < len(raw); i += size {
		var buf [size]byte
		copy(buf[:], raw[i:])
		data = append(data, int(int64(binary.LittleEndian.Uint64(buf[:]))))
	}

	return data
}

// intsToBytes — обратное преобразование, нужно только для seed-корпуса.
func intsToBytes(data []int) []byte {
	raw := make([]byte, 0, len(data)*8)
	for _, value := range data {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(int64(value)))
		raw = append(raw, buf[:]...)
	}

	return raw
}

// FuzzMaximum — дифференциальный фаззинг: для произвольных слайсов
// (включая отрицательные значения и любые длины) результаты maximum,
// maxChunks и эталонного slices.Max обязаны совпадать.
func FuzzMaximum(f *testing.F) {
	for _, seed := range [][]int{
		{1, 2, 3},
		{},
		{-5, -1, -10},
		{math.MinInt, 0, math.MaxInt},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 2}, // максимум в хвосте, длина не кратна CHUNKS
	} {
		f.Add(intsToBytes(seed))
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		data := bytesToInts(raw)
		want := maximum(data)

		if got := maxChunks(data); got != want {
			t.Fatalf("maxChunks(%v) = %d, maximum = %d", data, got, want)
		}
		if len(data) > 0 {
			if got := slices.Max(data); got != want {
				t.Fatalf("slices.Max(%v) = %d, maximum = %d", data, got, want)
			}
		}
	})
}

func BenchmarkMaximum(b *testing.B) {
	data := generateRandomElements(10_000_000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		maximum(data)
	}
}

func BenchmarkMaxChunks(b *testing.B) {
	data := generateRandomElements(10_000_000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		maxChunks(data)
	}
}
