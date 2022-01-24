package main

import (
	"bufio"
	"os"

	"github.com/emil14/neva/internal/new/core"
)

func Read(io core.IO) error {
	out, err := io.Out.Port("out")
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			out <- core.NewStrMsg(scanner.Text())
		}
	}()

	return nil
}
