package libcloser

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const defaultGracefulTimeout = 5

// GracefulTimeout seconds to wait until forced shutdown.
type GracefulTimeout string

// Closer provides abstraction on registered closer funcs and interfaces to handle them upon the OS signals.
type Closer struct {
	sync.Mutex
	ctx             context.Context
	once            sync.Once
	closers         []io.Closer
	closeFuncs      []func()
	sigCh           chan os.Signal
	sigs            []os.Signal
	gracefulTimeout time.Duration
	done            chan struct{}
	logger          *zap.Logger
}

// NewCloser construct a new Closer object.
func NewCloser(
	ctx context.Context,
	gracefulTimeout GracefulTimeout,
) *Closer {
	d, err := strconv.Atoi(string(gracefulTimeout))
	if err != nil {
		d = defaultGracefulTimeout
	}

	closer := Closer{
		ctx:             ctx,
		sigs:            []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		sigCh:           make(chan os.Signal, 1),
		gracefulTimeout: time.Duration(d) * time.Second,
		done:            make(chan struct{}, 1),
		logger:          zap.L().Named("closer"),
	}

	go func() {
		signal.Notify(closer.sigCh, closer.sigs...)
		sig := <-closer.sigCh
		closer.logger.Info("received syscall signal", zap.String("signal", sig.String()))
		closer.drop()
	}()

	go func() {
		<-ctx.Done()
		closer.logger.Info("received context cancellation")
		closer.drop()
	}()

	return &closer
}

// AddCloser any io.Closer
func (s *Closer) AddCloser(cl io.Closer) {
	s.Lock()
	defer s.Unlock()
	s.closers = append(s.closers, cl)
}

// AddFunc any func() that will be run on shutdown.
func (s *Closer) AddFunc(f func()) {
	s.Lock()
	defer s.Unlock()
	s.closeFuncs = append(s.closeFuncs, f)
}

// Close jobs forcibly, bypassing system signals
func (s *Closer) Close() {
	s.drop()
}

// Wait for job closing completion signal
func (s *Closer) Wait() {
	<-s.done
}

// GetWaitChan get a channel to receive closing completion signal
func (s *Closer) GetWaitChan() <-chan struct{} {
	return s.done
}

func (s *Closer) drop() {
	s.once.Do(func() {
		s.Lock()
		defer s.Unlock()

		// prepare sync mechanism
		wg := sync.WaitGroup{}
		chGracefulJobClose := make(chan struct{})
		chTimeoutJobClose := time.After(s.gracefulTimeout)

		for _, cl := range s.closers {
			wg.Add(1)
			go func(cl io.Closer) {
				defer wg.Done()
				err := cl.Close()
				if err != nil {
					s.logger.Error("failed to close", zap.Error(err))
				}
			}(cl)
		}

		for _, cf := range s.closeFuncs {
			wg.Add(1)
			go func(cf func()) {
				defer wg.Done()
				cf()
			}(cf)
		}

		// run wait after all the closers startup, preventing empty wg
		go func() {
			wg.Wait()
			close(chGracefulJobClose)
		}()

		s.logger.Info(
			"waiting before terminate or end up earlier if funcs are ready",
			zap.Float64("timeout seconds", s.gracefulTimeout.Seconds()),
			zap.Int("jobs to close amount", len(s.closers)+len(s.closeFuncs)),
		)

		select {
		case <-chGracefulJobClose:
			s.logger.Info("jobs are gracefully closed")
			s.done <- struct{}{}
		case <-chTimeoutJobClose:
			s.logger.Info("jobs are not closed due to timeout")
			s.done <- struct{}{}
		}
	})
}
