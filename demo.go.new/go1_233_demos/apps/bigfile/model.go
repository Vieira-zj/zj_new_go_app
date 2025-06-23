package main

import "sync"

type Chunk struct {
	payloads []*Payload
}

type Payload struct {
	Data []byte
}

var payloadPool = sync.Pool{
	New: func() any {
		return &Payload{}
	},
}

func getPayload() *Payload {
	return payloadPool.Get().(*Payload)
}

func releasePayload(payload *Payload) {
	payload.Data = payload.Data[:0]
	payloadPool.Put(payload)
}

var chunkPool = sync.Pool{
	New: func() any {
		return &Chunk{
			payloads: make([]*Payload, 0),
		}
	},
}

func getChunk() *Chunk {
	return chunkPool.Get().(*Chunk)
}

func releaseChunk(chunk *Chunk) {
	chunk.payloads = chunk.payloads[:0]
	chunkPool.Put(chunk)
}
