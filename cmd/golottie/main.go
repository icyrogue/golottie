package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/galihrivanto/go-inkscape"
	"github.com/icyrogue/golottie"
)

const (
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

	logger := newLogger(opts.quiet)
	logger.Debug(*opts)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(math.MaxInt))
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
	conv := newConverter(&wg, input)
	for i := 0; i < opts.workers; i++ {
		go conv.run(ctx, opts.output)
	}
	frame := frame{
		width:  opts.width,
		height: opts.height,
	}
	for renderer.NextFrame() {
		var buf string
		err := renderer.RenderFrameSVG(&buf)
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
}

type frame struct {
	buf    string
	num    int
	width  int
	height int
}

func newConverter(wg *sync.WaitGroup, input chan frame) *converter {
	return &converter{
		wg:    wg,
		input: input,
	}
}

//gocyclo:ignore
func (c *converter) run(ctx golottie.Context, output string) {
	c.wg.Add(1)
	// frames := make([]frame, 16)
	// var i int
	// TODO: add the option for verbose output
	proxy := inkscape.NewProxy(inkscape.Verbose(true))
	if err := proxy.Run(); err != nil {
		ctx.Error(err)
	}
	defer proxy.Close()
	render := func(v frame) error {
		f, err := os.CreateTemp(os.TempDir(), fmt.Sprintf(`%d-%d.svg`, time.Now().Unix(), v.num))
		if err != nil {
			return err
		}
		if len(v.buf) == 0 {
			return nil
		}
		_, err = f.WriteString(v.buf)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = proxy.RawCommands(
			"file-open:"+f.Name(),
			"export-filename:"+fmt.Sprintf(output, v.num),
			"export-do",
			"file-close",
		)
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
			// frames[i] = v
			// i++
			// if i != cap(frames) {
			// 	log.Info(i)
			// 	continue loop
			// }
			// for _, v := range frames {
			// if err := render(v); err != nil {
			// 	ctx.Error(err)
			// }
			// i = 0
			// }
		}
	}
	for {
		log.Warn("sweeping")
		if len(c.input) <= 0 {
			log.Warn("exiting")
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
