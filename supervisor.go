package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Service interface {
	ID() string
	Run(context.Context) error
	Validate() error
	RestartAfterError() (time.Duration, bool)
}

type Supervisor struct {
	Services []Service
}

func (s *Supervisor) validate() (err error) {
	for _, service := range s.Services {
		err = service.Validate()
		if err != nil {
			return fmt.Errorf("validate[%s] %w", service.ID(), err)
		}
	}

	return
}

func (s *Supervisor) Execute(ctx context.Context) error {
	if err := s.validate(); err != nil {
		return err
	}

	var wg sync.WaitGroup

	var once sync.Once
	var errOnce error

	for _, service := range s.Services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					slog.Info(
						"Exiting...",
						slog.String(
							"id", s.ID(),
						),
					)

					return

				default:
				}

				slog.Debug(
					"Starting...",
					slog.String(
						"id", s.ID(),
					),
				)

				if err := s.Run(ctx); err != nil {
					slog.Error(
						"Error starting",
						slog.String(
							"id", s.ID(),
						),
					)

					if err := ctx.Err(); err != nil {
						once.Do(func() {
							errOnce = err
						})

						return
					}

					if restartAfter, ok := s.RestartAfterError(); ok {
						slog.Warn(
							"Waiting to restart",
							slog.String(
								"id", s.ID(),
							),
							slog.String(
								"restart_after", restartAfter.String(),
							),
						)

						time.Sleep(
							restartAfter,
						)

						continue
					}
				}

				return
			}
		}(service)
	}

	wg.Wait()

	return errOnce
}
