package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/icyrogue/golottie"
)

const (
	defWidth   = 1920
	defHeight  = 1080
	defBufSize = 16
	defWorkers = 1
)

func main() {
	opts := parseFlags()
	opts.flagSet.Parse(opts.args)
	if opts.input == "" || opts.output == "" {
		log.Fatal("--output or --input is not provided, try --help")
	}

	logger := newLogger(opts.quiet)
	logger.Debug(*opts)
	ctx := context.Background()
	run(ctx, logger, opts)
}

func run(ctx context.Context, logger log.Logger, opts *options) {
	logger.Info("Launching the browser")
	renderer, err := golottie.New(ctx)
	if err != nil {
		logger.Fatal(err.Error())
	}
	ctx = renderer.GetContext()

	logger.Info("Parsing animation", "file", opts.input)
	animation, err := golottie.AnimationFromFile(opts.input)
	if err != nil {
		logger.Fatal(err.Error())
	}
	renderer.SetAnimation(animation)

	logger.Info("Allocating frame buffer", "size", opts.bufSize)
	input := make(chan golottie.Frame, opts.bufSize)
	framesTotal := animation.GetFramesTotal()
	logger.Info("Starting converter", "frames", framesTotal)
	var wg sync.WaitGroup
	conv := newConverter(&wg, input)
	for i := 0; i < opts.workers; i++ {
		go conv.run(ctx, opts.output)
	}
	frame := golottie.Frame{
		Width:  opts.width,
		Height: opts.height,
	}
	for renderer.NextFrame() {
		var buf string
		err := renderer.RenderFrameSvg(&buf)
		if err != nil {
			log.Fatal(err.Error())
		}
		frame.Num++
		frame.Buf = &buf
		input <- frame

		//debug
	}
	if renderer.Error != nil {
		log.Fatal(err.Error())
	}
	renderer.Close()
	wg.Wait()
}

type converter struct {
	wg    *sync.WaitGroup
	input chan golottie.Frame
}

func newConverter(wg *sync.WaitGroup, input chan golottie.Frame) *converter {
	return &converter{
		wg:    wg,
		input: input,
	}
}

func (c *converter) run(ctx context.Context, output string) {
	c.wg.Add(1)
	render := func(v golottie.Frame) error {
		f, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%d.svg", time.Now().UnixMicro()))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(*v.Buf)
		if err != nil {
			return err
		}
		cmd := exec.Command("rsvg-convert", "-w", strconv.Itoa(v.Width),
			"-h", strconv.Itoa(v.Height), f.Name(), "-o", fmt.Sprintf(output, v.Num))
		var error []byte
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("convertion error (stderr: %s): %w, %s", error, err, cmd)
		}
		return nil
	}
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case v := <-c.input:
			fmt.Printf("\r---> Rendering frame %d", v.Num)
			render(v)
		}
	}
	for {
		if len(c.input) <= 0 {
			break
		}
		v := <-c.input
		fmt.Printf("\r---> Rendering frame %d", v.Num)
		if err := render(v); err != nil {
			log.Fatal(err.Error())
		}
	}
	c.wg.Done()
}

type options struct {
	width  int
	height int

	input  string
	output string

	quiet   bool
	workers int
	bufSize int

	flagSet flag.FlagSet
	args    []string
}

func parseFlags() *options {
	opts := options{
		width:   defWidth,
		height:  defHeight,
		bufSize: defBufSize,
		workers: defWorkers,

		flagSet: *flag.CommandLine,
		args:    os.Args[1:],
	}
	opts.flagSet.StringVar(&opts.input, "input", "", "input file name")
	opts.flagSet.StringVar(&opts.input, "i", "", "")
	opts.flagSet.StringVar(&opts.output, "output", "", "output sprintf pattern")
	opts.flagSet.StringVar(&opts.output, "o", "", "Ex: render/%04d.png")
	opts.flagSet.IntVar(&opts.width, "width", defWidth, "width of the output")
	opts.flagSet.IntVar(&opts.width, "w", defWidth, "")
	opts.flagSet.IntVar(&opts.height, "height", defHeight, "height of the output")
	opts.flagSet.IntVar(&opts.height, "h", defHeight, "")
	opts.flagSet.IntVar(&opts.workers, "count", defWorkers, "worker count (goroutines) to be created for concurrent rendering")
	opts.flagSet.IntVar(&opts.workers, "c", defWorkers, "")
	opts.flagSet.IntVar(&opts.bufSize, "bufsize", defBufSize, "frame buffer size")
	opts.flagSet.IntVar(&opts.bufSize, "b", defBufSize, "short for --bufsize")
	opts.flagSet.BoolVar(&opts.quiet, "quiet", false, "should I have a mouth to scream?")
	opts.flagSet.BoolVar(&opts.quiet, "q", false, "")
	opts.flagSet.Usage = func() {
		fmt.Fprint(opts.flagSet.Output(), "Usage of golottie:\n\n")
		var b strings.Builder
		opts.flagSet.VisitAll(func(f *flag.Flag) {
			if len(f.Name) == 1 {
				fmt.Fprintf(&b, "-%s ", f.Name)
				return
			}
			fmt.Fprintf(&b, "--%s\t%s\n", f.Name, f.Usage)
			if f.DefValue != "" {
				fmt.Fprintf(&b, "\t\t(default: %s)\n", f.DefValue)
			}
		})
		fmt.Fprint(opts.flagSet.Output(), b.String())
	}

	return &opts
}

func newLogger(quiet bool) log.Logger {
	logger := log.New()
	if quiet {
		logger.SetLevel(log.FatalLevel)
	}
	return logger
}
