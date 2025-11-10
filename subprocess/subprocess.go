package subprocess

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

type Subprocess struct {
	Name string
	Cmd  []string
}

func (s *Subprocess) ID() string {
	return s.Name
}

func (s *Subprocess) RestartAfterError() (time.Duration, bool) {
	return 5 * time.Second, true
}

func (s *Subprocess) Validate() error {
	if len(s.Cmd) == 0 {
		return errors.New("no command was given")
	}

	return nil
}

func (s *Subprocess) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, s.Cmd[0], s.Cmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}
