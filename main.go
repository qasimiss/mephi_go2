package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Количество элементов в слайсе случайных чисел.
const SIZE = 100_000_000

// Количество частей, на которые maxChunks делит слайс.
const CHUNKS = 8

// generateRandomElements возвращает слайс из size случайных неотрицательных целых чисел. Если size меньше или равен нулю, возвращает пустой слайс.
func generateRandomElements(size int) []int {
	if size <= 0 {
		return []int{}
	}

	elements := make([]int, size)
	for i := range elements {
		elements[i] = rand.Int()
	}

	return elements
}

// maximum возвращает максимальное значение слайса data.
// Для пустого слайса возвращает 0 — по условию задачи элементы неотрицательные поэтому 0 однозначно означает что элементов нет.
func maximum(data []int) int {
	if len(data) == 0 {
		return 0
	}

	maxValue := data[0]
	for _, value := range data[1:] {
		if value > maxValue {
			maxValue = value
		}
	}

	return maxValue
}

// maxChunks ищет максимум параллельно: делит слайс на части (CHUNKS штук, но не больше, чем элементов), находит максимум каждой части в отдельной горутине и возвращает максимум из найденных значений.
// Для пустого слайса, как и maximum, возвращает 0.
func maxChunks(data []int) int {
	if len(data) == 0 {
		return 0
	}

	chunks := min(CHUNKS, len(data))
	chunkSize := len(data) / chunks

	// Каждая горутина пишет в собственную ячейку maxima, поэтому рейса нет;
	// wg.Wait() гарантирует, что все записи завершены до финального прохода.
	maxima := make([]int, chunks)

	var wg sync.WaitGroup
	wg.Add(chunks)

	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == chunks-1 {
			end = len(data) // последний кусок забирает остаток деления
		}

		go func(chunk []int, index int) {
			defer wg.Done()
			maxima[index] = maximum(chunk)
		}(data[start:end], i)
	}

	wg.Wait()

	return maximum(maxima)
}

func main() {
	fmt.Printf("Генерируем %d целых чисел\n", SIZE)
	data := generateRandomElements(SIZE)

	fmt.Println("Ищем максимальное значение в один поток")
	singleStart := time.Now()
	singleMax := maximum(data)
	singleElapsed := time.Since(singleStart).Microseconds()
	fmt.Printf("Максимальное значение элемента: %d\nВремя поиска: %d мкс\n\n", singleMax, singleElapsed)

	fmt.Printf("Ищем максимальное значение в %d потоков\n", CHUNKS)
	chunksStart := time.Now()
	chunksMax := maxChunks(data)
	chunksElapsed := time.Since(chunksStart).Microseconds()
	fmt.Printf("Максимальное значение элемента: %d\nВремя поиска: %d мкс\n", chunksMax, chunksElapsed)
}
