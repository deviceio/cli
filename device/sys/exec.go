package sys

import (
	"context"
	"io"
	"log"
	"os"

	sdk "github.com/deviceio/sdk/go-sdk"
)

func Exec(deviceid, cmd string, args []string, c sdk.Client) {
	ctx, cancel := context.WithCancel(context.Background())

	process, err := c.Device(deviceid).Process().Create(ctx, cmd, args)

	if err != nil {
		panic(err.Error())
	}

	defer func() {
		err := process.Delete(ctx)
		if err != nil {
			log.Fatal("error destroying process:", err.Error())
		}
	}()

	done := make(chan bool)
	stdin := process.Stdin(ctx)
	stdout := process.Stdout(ctx)
	stderr := process.Stderr(ctx)

	go func() {
		if _, err := io.Copy(os.Stdout, stdout); err != nil {
			if err != io.EOF {
				cancel()
			}
		}
		done <- true
	}()

	go func() {
		if _, err := io.Copy(os.Stderr, stderr); err != nil {
			if err != io.EOF {
				cancel()
			}
		}
		done <- true
	}()

	go func() {
		if _, err := io.Copy(stdin, os.Stdin); err != nil {
			if err != io.EOF {
				cancel()
			}
		}
		done <- true
	}()

	err = process.Start(ctx)

	if err != nil {
		cancel()
	}

	select {
	case <-ctx.Done():
		log.Fatal(ctx.Err())
	case <-done:
	}
}
