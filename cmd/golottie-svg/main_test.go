package main

import (
	"context"
	"testing"
)

func BenchmarkMain(b *testing.B) {
	opts := options{
		input:   "../../misc/test.json",
		output:  "../../misc/render2/%04d.png",
		width:   600,
		height:  600,
		workers: 2,
	}
	logger := newLogger(true)
	for i := 0; i < b.N; i++ {
		run(context.Background(), logger, &opts)
	}
}
