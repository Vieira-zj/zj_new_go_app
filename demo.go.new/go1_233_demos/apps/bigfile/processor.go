package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

func processSingleFile(file io.Reader, numWorkers int, chunkSize int) {
	scanner := bufio.NewScanner(file)
	chunkChan := make(chan *Chunk, numWorkers)

	wg := sync.WaitGroup{}
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunkChan {
				for _, payload := range chunk.payloads {
					// simulate processing time of each payload
					time.Sleep(time.Second)
					fmt.Println(string(payload.Data))
					releasePayload(payload)
				}
				releaseChunk(chunk)
			}
		}()
	}

	go func() {
		defer close(chunkChan)
		for {
			chunk := getChunk()
			for len(chunk.payloads) < chunkSize && scanner.Scan() {
				payload := getPayload()
				payload.Data = payload.Data[:0]
				payload.Data = append(payload.Data[:0], scanner.Bytes()...)
				chunk.payloads = append(chunk.payloads, payload)
			}

			if len(chunk.payloads) > 0 {
				chunkChan <- chunk
			} else {
				releaseChunk(chunk)
				break
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println("Scanner file:", err)
		}
	}()

	wg.Wait()
}
