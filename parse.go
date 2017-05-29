package main

import (
	"bufio"
	"context"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

var fpingArgs = []string{"-B 1", "-D", "-r0", "-O 0", "-Q 10", "-p 1000", "-l"}

// Point represents the fping results for a single host
type Point struct {
	Host        string
	LossPercent int
	Min         float64
	Avg         float64
	Max         float64
}

func runAndRead(ctx context.Context, hosts []string, con Client) error {
	args := append([]string(nil), fpingArgs...)
	for _, v := range hosts {
		args = append(args, v)
	}
	cmd, err := exec.LookPath("fping")
	if err != nil {
		return err
	}
	runner := exec.Command(cmd, args...)
	stderr, err := runner.StderrPipe()
	if err != nil {
		return err
	}
	runner.Start()

	buff := bufio.NewScanner(stderr)
	for buff.Scan() {
		text := buff.Text()
		fields := strings.Fields(text)

		if len(fields) > 1 {
			host := fields[0]
			data := fields[4]
			dataSplitted := strings.FieldsFunc(data, slashSplitter)
			// Remove ,
			dataSplitted[2] = strings.TrimRight(dataSplitted[2], "%,")
			lossp := mustInt(dataSplitted[2])
			min, max, avg := 0.0, 0.0, 0.0
			// Ping times
			if len(fields) > 5 {
				times := fields[7]
				td := strings.FieldsFunc(times, slashSplitter)
				min, avg, max = mustFloat(td[0]), mustFloat(td[1]), mustFloat(td[2])
			}
			pt := Point{Host: host, Min: min, Max: max, Avg: avg, LossPercent: lossp}
			if err := con.Write(pt); err != nil {
				log.Printf("Error writing data point: %s", err)
			}
		}
	}
	return nil
}

func mustInt(data string) int {
	in, err := strconv.Atoi(data)
	if err != nil {
		return 0
	}
	return in
}

func mustFloat(data string) float64 {
	flt, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return 0.0
	}
	return flt
}

func slashSplitter(c rune) bool {
	return c == '/'
}