// mark command line tool. available at https://github.com/a8m/mark
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/a8m/mark"
)

var (
	input     = flag.String("i", "", "")
	output    = flag.String("o", "", "")
	bwsr      = flag.Bool("b", false, "")
	port      = flag.String("port", "8080", "")
	smarty    = flag.Bool("smartypants", false, "")
	fractions = flag.Bool("fractions", false, "")
)

var usage = `Usage: mark [options...] <input>

Options:
  -i  Specify file input, otherwise use last argument as input file. 
      If no input file is specified, read from stdin.
  -o  Specify file output. If none is specified, write to stdout.

  -b     Open output in the browser.
  -port  Port for browser flag. default to 8080.

  -smartypants  Use "smart" typograhic punctuation for things like 
                quotes and dashes.
  -fractions    Traslate fraction like to suitable HTML elements
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage))
	}
	flag.Parse()
	// read
	var reader *bufio.Reader
	if *input != "" {
		file, err := os.Open(*input)
		if err != nil {
			usageAndExit(fmt.Sprintf("Error to open file input: %s.", *input))
		}
		defer file.Close()
		reader = bufio.NewReader(file)
	} else {
		stat, err := os.Stdin.Stat()
		if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
			usageAndExit("")
		}
		reader = bufio.NewReader(os.Stdin)
	}
	// collect data
	var data string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			usageAndExit("failed to reading input.")
		}
		data += line
	}
	// write
	var (
		err  error
		file = os.Stdout
	)
	// mark rendering
	opts := mark.DefaultOptions()
	opts.Smartypants = *smarty
	opts.Fractions = *fractions
	if *bwsr {
		b := &browser{
			port: *port,
			path: *input,
			parseFunc: func(s string) string {
				return mark.New(s, opts).Render()
			},
		}
		go b.watch()
		b.Serve()
	}
	m := mark.New(data, opts)
	if *output != "" {
		file, err = os.Create(*output)
		failOnErr(err, "create the output faile")
	}
	_, err = file.WriteString(m.Render())
	failOnErr(err, fmt.Sprintf("write output to: %s", file.Name()))
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("mark: failed to %s: %s", msg, err)
	}
}
