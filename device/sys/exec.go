package sys

import (
	"context"
	"io"
	"log"
	"os"

	"sync"

	sdk "github.com/deviceio/sdk/go-sdk"
)

func Exec(deviceid, cmd string, args []string, c sdk.Client) {
	ctx, cancel := context.WithCancel(context.Background())

	process, err := c.Device(deviceid).Process().Create(ctx, cmd, args)

	if err != nil {
		panic(err.Error())
	}

	defer func() {
		err := process.Delete(context.Background())
		if err != nil {
			log.Println("error destroying process:", err.Error())
		}
	}()

	data := &sync.WaitGroup{}
	done := make(chan bool)
	stdin := process.Stdin(ctx)
	stdout := process.Stdout(ctx)
	stderr := process.Stderr(ctx)

	stdoutbuf := make([]byte, 250000)
	stderrbuf := make([]byte, 250000)
	stdinbuf := make([]byte, 250000)

	go func() {
		data.Add(1)
		for {
			n, err := stdout.Read(stdoutbuf)

			if n > 0 {
				os.Stdout.Write(stdoutbuf[:n])
			}

			if err != nil {
				if err != io.EOF {
					cancel()
				}

				break
			}
		}
		data.Done()
	}()

	go func() {
		data.Add(1)
		for {
			n, err := stderr.Read(stderrbuf)

			if n > 0 {
				os.Stderr.Write(stderrbuf[:n])
				os.Stderr.Sync()
			}

			if err != nil {
				if err != io.EOF {
					cancel()
				}

				break
			}
		}
		data.Done()
	}()

	go func() {
		for {
			n, err := os.Stdin.Read(stdinbuf)

			if n > 0 {
				stdin.Write(stdinbuf[:n])
			}

			if err != nil {
				if err != io.EOF {
					cancel()
				}

				break
			}
		}
	}()

	err = process.Start(ctx)

	if err != nil {
		cancel()
	}

	go func() {
		data.Wait()
		done <- true
	}()

	select {
	case <-ctx.Done():
		log.Fatal(ctx.Err())
	case <-done:
		return
	}
}
