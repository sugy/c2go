package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	ExitCodeOK = iota
	ExitCodeParseError
	ExitCodeParseFlagError
)

type CLI struct {
	inStream             io.Reader
	outStream, errStream io.Writer
}

type Line struct {
	Spacer string
	Count  uint64
	Text   string
	Graph  string
}

func (c *CLI) Run(args []string) int {
	var version bool
	var max uint64
	var total uint64
	var totalFlag bool
	var target uint64
	var length int
	var verbose bool

	flags := flag.NewFlagSet("c2g", flag.ContinueOnError)
	flags.SetOutput(c.errStream)
	flags.BoolVar(&version, "version", false, "Print version information and quit")
	flags.BoolVar(&totalFlag, "total", false, "Take a percentage from the total.")
	flags.IntVar(&length, "length", 30, "Specify length of Graph.")
	flags.BoolVar(&verbose, "verbose", false, "Verbose mode.")

	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeParseFlagError
	}

	if version {
		fmt.Fprintf(c.errStream, "c2g version %s\n", Version)
		return ExitCodeOK
	}

	lines := make([]Line, 0, 1024)
	if err := Parser(c.inStream, &lines); err != nil {
		return ExitCodeParseError
	}

	for _, i := range lines {
		if max < i.Count {
			max = i.Count
		}
		total += i.Count
	}

	switch {
	case length > 100:
		length = 100
	case length < 10:
		length = 10
	}

	if totalFlag {
		target = total
	} else {
		target = max
	}

	GenGraph(lines, target, length, verbose)

	PrintLine(c.outStream, lines)

	return ExitCodeOK
}

func Parser(stdin io.Reader, lines *[]Line) error {
	scanner := bufio.NewScanner(stdin)
	rep := regexp.MustCompile(`^(\s*)([0-9]+) (.*)$`)
	for scanner.Scan() {
		line := rep.FindSubmatch([]byte(scanner.Text()))
		spacer := string(line[1])
		count, err := strconv.ParseUint(string(line[2]), 10, 64)
		if err != nil {
			return fmt.Errorf("%s", "Parse Error")
		}
		text := string(line[3])

		*lines = append(*lines, Line{Spacer: spacer, Count: count, Text: text})
	}

	return nil
}

func GenGraph(lines []Line, target uint64, length int, verbose bool) {

	n := len(lines)
	if verbose {
		for i := 0; i < n; i++ {
			barCount := int(lines[i].Count * uint64(length) / target)
			percentage := int(lines[i].Count * 100 / target)
			charLen := len(strconv.Itoa(percentage))
			var format strings.Builder
			format.WriteString("[%s%")
			format.WriteString(strconv.Itoa(charLen))
			format.WriteString("d]")
			var graph strings.Builder
			graph.WriteString(strings.Repeat("|", barCount))
			graph.WriteString(strings.Repeat(" ", length-barCount))
			lines[i].Graph = fmt.Sprintf(format.String(), graph.String()[:length-len(strconv.Itoa(percentage))], percentage)
		}
	} else {
		for i := 0; i < n; i++ {
			barCount := int(lines[i].Count * uint64(length) / target)
			var graph strings.Builder
			graph.WriteString(strings.Repeat("|", barCount))
			graph.WriteString(strings.Repeat(" ", length-barCount))
			lines[i].Graph = fmt.Sprintf("[%s]", graph.String())
		}
	}
}

func PrintLine(stdout io.Writer, lines []Line) {

	for _, line := range lines {
		fmt.Fprintf(stdout, "%s%d %s %s\n",
			line.Spacer, line.Count, line.Graph, line.Text)
	}
}
