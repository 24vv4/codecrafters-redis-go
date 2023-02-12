package main

import "math"
import "time"

type record struct {
    value string
    timestamp int64
}

type Memory struct {
    data map[string]record
}

func NewMemory() *Memory {
    return &Memory {
        data: make(map[string]record),
    }
}

func (kv *Memory) Set(key, value string) {
    kv.data[key] = record{value: value, timestamp: math.MaxInt32}
}

func (kv *Memory) SetPX(key, value string, ms int64) {
    kv.data[key] = record{value: value, timestamp: time.Now().Unix() + ms}

}

func (kv *Memory) Get(key string) (string, bool) {
    rec := kv.data[key]
    if(rec.timestamp > time.Now().Unix()) {
        return "", false
    } else {
        return kv.data[key].value, true
    }
}
