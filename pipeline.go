package execpipe

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

var logger *slog.Logger

type Executor struct {
	opts *ExecutorOptions
}

type ExecutorOptions struct {
	Logger *slog.Logger
}

func NewExecutor() Executor {
	logger = defaultOptions().Logger
	return Executor{opts: defaultOptions()}
}

func NewExecutorWithOptions(opts *ExecutorOptions) Executor {
	logger = opts.Logger
	return Executor{
		opts: opts,
	}
}

func defaultOptions() *ExecutorOptions {
	return &ExecutorOptions{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

type Command struct {
	Executable string
	Args       []string
	Env        map[string]string
}

func (c Command) isEmpty() bool {
	return c.Executable == "" && len(c.Args) == 0 && len(c.Env) == 0
}

// RunR executes a sequence of commands connected via pipes in the provided order and returns the output reader of the last command.
// It takes a variadic number of Command arguments, where each Command defines an executable, its arguments, and environment variables.
// Commands are executed sequentially, with the standard output of each being piped to the standard input of the next.
// Empty commands are skipped.
// If any command fails to start or an error occurs while setting up pipes, it returns an error.
// The stderr stream of all commands is redirected to os.Stderr for error visibility.
func (e Executor) RunR(commands ...Command) (io.ReadCloser, error) {
	logCommands(commands)
	var err error
	var cmds []*exec.Cmd

	for i, command := range commands {
		if cmd := command.prepareCmd(); cmd != nil {
			cmds = append(cmds, cmd)
		} else {
			return nil, fmt.Errorf("empty command")
		}
		if i > 0 {
			if cmds[i].Stdin, err = cmds[i-1].StdoutPipe(); err != nil {
				return nil, err
			}
		}
		cmds[i].Stderr = os.Stderr
	}
	out, err := cmds[len(cmds)-1].StdoutPipe()
	if err != nil {
		return nil, err
	}
	for _, c := range cmds {
		slog.Debug("starting command", "path", c.Path)
		if err = c.Start(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// RunE runs a series of Command objects sequentially, piping the stdout of each into the stdin of the next, and returns an error if any command fails.
func (e Executor) RunE(commands ...Command) error {
	logCommands(commands)
	var err error
	var cmds []*exec.Cmd

	for i, command := range commands {
		if cmd := command.prepareCmd(); cmd != nil {
			cmds = append(cmds, cmd)
		} else {
			return fmt.Errorf("empty command")
		}
		if i > 0 {
			if cmds[i].Stdin, err = cmds[i-1].StdoutPipe(); err != nil {
				return err
			}
		}
		cmds[i].Stderr = os.Stderr
	}
	for _, c := range cmds {
		slog.Debug("starting command", "command", c.Path)
		if err = c.Start(); err != nil {
			return err
		}
	}
	for _, c := range cmds {
		slog.Debug("waiting for command", "command", c.Path)
		err = c.Wait()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Command) prepareCmd() *exec.Cmd {
	if c.isEmpty() {
		return nil
	}
	cmd := exec.Command(c.Executable, c.Args...)
	cmd.Env = c.convertEnv()
	return cmd
}

func (c Command) convertEnv() []string {
	var res []string
	for key, val := range c.Env {
		if key != "" {
			res = append(res, fmt.Sprintf("%s=%s", key, val))
		}
	}
	return res
}

func logCommands(cmds []Command) {
	slog.Debug(func() string {
		names := strings.Builder{}
		for i, c := range cmds {
			names.WriteString(c.String())
			if i < len(cmds)-1 {
				names.WriteString(" | ")
			}
		}
		return names.String()
	}(),
	)
}

func (c Command) String() string {
	envBuilder := strings.Builder{}
	for key, val := range c.Env {
		envBuilder.WriteString(fmt.Sprintf("%s='%s' ", key, val))
	}

	// TODO: Consider adding optional return arguments for commands to handle cases where output isn't needed directly.
	// TODO: Explore making the Env map optional or allowing overriding system environment variables.
	return c.Executable
}
