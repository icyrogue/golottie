package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/galihrivanto/go-inkscape"
	"github.com/icyrogue/golottie"
)

const (
	defTimeout = 1800
	defWidth   = 1920
	defHeight  = 1080
	defBufSize = 16
	defWorkers = 1
)

//gocyclo:ignore
func main() {
	opts := parseFlags()
	err := opts.flagSet.Parse(opts.args)
	if err != nil {
		log.Fatal(err)
	}
	if opts.input == "" || opts.output == "" {
		log.Fatal("--output or --input is not provided, try --help")
	}

	logger := newLogger(opts.verbose)
	logger.Warn(*opts)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opts.timeout)*time.Second)
	defer cancel()
	run(ctx, logger, opts)
}

//gocyclo:ignore
func run(ctxParent context.Context, logger log.Logger, opts *options) {
	logger.Info("Launching the browser")

	ctx, cancel := golottie.NewContext(ctxParent)
	renderer := golottie.New(ctx)
	logger.Info("Parsing animation", "file", opts.input)
	a, err := os.ReadFile(opts.input)
	if err != nil {
		logger.Fatal(err)
	}
	animation, err := golottie.NewAnimation(a).WithDefaultTemplate()
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = renderer.SetAnimation(animation)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Allocating frame buffer", "size", opts.bufSize)
	input := make(chan frame, opts.bufSize)
	framesTotal := animation.GetFramesTotal()
	logger.Info("Starting converter", "frames", framesTotal)
	var wg sync.WaitGroup
	conv := newConverter(&wg, input, opts)
	for i := 0; i < opts.workers; i++ {
		go conv.run(ctx, opts.output)
	}
	frame := frame{
		width:  opts.width,
		height: opts.height,
	}
	for renderer.NextFrame() {
		var buf []byte
		err := renderer.RenderFrame(&buf)
		if err != nil {
			log.Fatal(err.Error())
		}
		frame.num++
		frame.buf = buf
		input <- frame
	}
	if err = ctx.Errors[len(ctx.Errors)-1]; err != nil {
		if err != golottie.EOF {
			log.Fatal(err.Error())
		}
	}
	cancel()
	wg.Wait()
	logger.Info("Done!", "output", path.Dir(opts.output))
}

type converter struct {
	wg    *sync.WaitGroup
	input chan frame
	opts  *options
}

type frame struct {
	buf    []byte
	num    int
	width  int
	height int
}

func newConverter(wg *sync.WaitGroup, input chan frame, opts *options) *converter {
	return &converter{
		opts:  opts,
		wg:    wg,
		input: input,
	}
}

//gocyclo:ignore
func (c *converter) run(ctx golottie.Context, output string) {
	c.wg.Add(1)
	proxy := inkscape.NewProxy(inkscape.Verbose(c.opts.verbose))
	if err := proxy.Run(); err != nil {
		ctx.Error(err)
	}
	defer proxy.Close()
	render := func(v frame) error {
		err := os.WriteFile(fmt.Sprintf(output, v.num), v.buf, 7777)
		if err != nil {
			return err
		}
		return nil
	}
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case v := <-c.input:
			fmt.Printf("\r---> Rendering frame %d", v.num)
			if err := render(v); err != nil {
				ctx.Error(err)
			}
		}
	}
	for {
		if len(c.input) <= 0 {
			break
		}
		v := <-c.input
		fmt.Printf("\r---> Rendering frame %d", v.num)
		if err := render(v); err != nil {
			ctx.Error(err)
		}
	}
	fmt.Printf("\r")
	c.wg.Done()
}

type options struct {
	width  int
	height int

	input  string
	output string

	verbose bool
	workers int
	bufSize int
	timeout int

	flagSet flag.FlagSet
	args    []string
}

func parseFlags() *options {
	opts := options{
		width:   defWidth,
		height:  defHeight,
		bufSize: defBufSize,
		workers: defWorkers,
		timeout: defTimeout,
		verbose: false,

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
	opts.flagSet.BoolVar(&opts.verbose, "verbose", true, "should I have a mouth to scream?")
	opts.flagSet.BoolVar(&opts.verbose, "q", false, "")

	if t, err := strconv.Atoi(os.Getenv("GOLOTTIE_TIMEOUT")); err != nil && opts.verbose {
		log.Warn("setting timeout value to default:", err)
	} else if t != 0 {
		opts.timeout = t
	}
	log.Warn(opts.timeout)
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

func newLogger(verbose bool) log.Logger {
	logger := log.New()
	if verbose {
		logger.SetLevel(log.FatalLevel)
	}
	return logger
}
