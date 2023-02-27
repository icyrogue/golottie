package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/icyrogue/golottie"
)

//gocyclo:ignore
func main() {
	// Initialize a new animation from file
	a, err := os.ReadFile("../../misc/test.json")
	if err != nil {
		log.Fatal(err)
	}

	// Use default template to create HTML needed for rendering
	animation, err := golottie.NewAnimation(a).WithDefaultTemplate()
	if err != nil {
		log.Fatal(err)
	}

	// NewContext creates a new chromedp context with an error stack
	ctx, cancel := golottie.NewContext(context.Background())
	defer cancel()

	// Create a new render instance from context and set the animation to render
	instance := golottie.New(ctx)
	err = instance.SetAnimation(animation)
	if err != nil {
		log.Fatal(err)
	}

	// Advance the current frame and render it as PNG
	var frame int
	for instance.NextFrame() {
		//	frame++
		log.Println("Rendering frame", frame)
		// The render result will be stored in the frame buffer
		var buf []byte
		err := instance.RenderFrame(&buf)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(fmt.Sprintf("../render/%04d.png", frame), buf, 0o644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Check the context error stack
	if err = ctx.Errors[:1][0]; err != nil {
		// Instance will return EOF at the end of animation frames
		if err != golottie.EOF {
			log.Fatal(err)
		}
	}
}
