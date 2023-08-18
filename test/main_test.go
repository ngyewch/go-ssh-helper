package test

import (
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"
)

var (
	testEnv = &TestEnv{}
)

func TestMain(m *testing.M) {
	testEnv0, err := NewTestEnv()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	testEnv = testEnv0

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println()
			fmt.Println("Interrupted by user")
			_ = testEnv.Close()
			os.Exit(1)
		}
	}()

	err = testEnv.Setup()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Waiting for containers to start...")
	time.Sleep(10 * time.Second)

	code := m.Run()

	err = testEnv.Close()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}
