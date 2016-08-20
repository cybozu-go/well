package cmd

import (
	"context"
	"errors"
	"testing"
)

func TestEnvironmentStop(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	waitCh := make(chan struct{})

	env.Go(func(ctx context.Context) error {
		return nil
	})
	env.Go(func(ctx context.Context) error {
		<-waitCh
		return nil
	})

	env.Stop()
	close(waitCh)
	err := env.Wait()
	if err != nil {
		t.Error(err)
	}
}

func TestEnvironmentError(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()

	testError := errors.New("test")

	stopCh := make(chan struct{})

	go func() {
		err := env.Wait()
		if err != testError {
			t.Error(`err != testError`)
		}
		close(stopCh)
	}()

	waitCh := make(chan struct{})

	env.Go(func(ctx context.Context) error {
		close(waitCh)
		return nil
	})

	<-waitCh
	env.Cancel(testError)

	<-stopCh
}

func TestEnvironmentGo(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()

	testError := errors.New("test")

	waitCh := make(chan struct{})

	env.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	env.Go(func(ctx context.Context) error {
		<-ctx.Done()
		// uncomment the next line delays test just 2 seconds.
		//time.Sleep(2 * time.Second)
		return nil
	})

	env.Go(func(ctx context.Context) error {
		<-waitCh
		return testError
	})

	close(waitCh)
	err := env.Wait()
	if err != testError {
		t.Error(`err != testError`)
	}
}
