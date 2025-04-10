package execpipe

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestCommand_convertEnv(t *testing.T) {
	tests := []struct {
		name     string
		Env      map[string]string
		wantFunc func(t *testing.T, got []string, err error)
	}{
		{
			name: "single env",
			Env:  map[string]string{"PGPASSWORD": "secret"},
			wantFunc: func(t *testing.T, got []string, err error) {
				assert.NoError(t, err)
				assert.Len(t, got, 1)
				assert.Contains(t, got, "PGPASSWORD=secret")
			},
		},
		{
			name: "multiple env",
			Env:  map[string]string{"NAME": "Darth", "LASTNAME": "Vader"},
			wantFunc: func(t *testing.T, got []string, err error) {
				assert.NoError(t, err)
				assert.Len(t, got, 2)
				assert.Contains(t, got, "NAME=Darth")
				assert.Contains(t, got, "LASTNAME=Vader")
			},
		},
		{
			name: "empty value",
			Env:  map[string]string{"NAME": ""},
			wantFunc: func(t *testing.T, got []string, err error) {
				assert.NoError(t, err)
				assert.Len(t, got, 1)
				assert.Contains(t, got, "NAME=")
			},
		},
		{
			name: "empty key",
			Env:  map[string]string{"": "Darth"},
			wantFunc: func(t *testing.T, got []string, err error) {
				assert.NoError(t, err)
				assert.Len(t, got, 0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Command{Env: tt.Env}
			tt.wantFunc(t, c.convertEnv(), nil)
		})
	}
}

func TestRunR(t *testing.T) {
	tests := []struct {
		name     string
		cmds     []Command
		wantFunc func(t *testing.T, got string, err error)
	}{
		{
			name: "single cmd",
			cmds: []Command{
				{
					Executable: "echo",
					Args:       []string{"-n", "hello"},
				},
			},
			wantFunc: func(t *testing.T, got string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "hello", got)
			},
		},
		{
			name: "contains empty command",
			cmds: []Command{
				{
					Executable: "echo",
					Args:       []string{"-n", "hello"},
				},
				{},
			},
			wantFunc: func(t *testing.T, got string, err error) {
				assert.Errorf(t, err, "empty command")
			},
		},
		{
			name: "multiple commands",
			cmds: []Command{
				{
					Executable: "echo",
					Args:       []string{"-n", "hello"},
				},
				{Executable: "cat"}, {Executable: "cat"}, {Executable: "cat"},
			},
			wantFunc: func(t *testing.T, got string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "hello", got)
			},
		},
		// {
		// 	name: "first failed",
		// 		cmds: []Command{
		// 			{
		// 				Executable: "sh",
		// 				Args:       []string{"-c", ">&2 echo \"error\";exit 1"},
		// 			},
		// 			{Executable: "wc", Args: []string{"-m"}},
		// 	},
		// 	wantOut: "",
		// 	wantErr: true,
		// },
		// {
		// 	name: "second failed",
		// 		cmds: []Command{
		// 			{
		// 				Executable: "echo",
		// 				Args:       []string{"first"},
		// 			},
		// 			{Executable: "sh", Args: []string{"-c", ">&2 echo \"error\";exit 1"}},
		// 	},
		// 	wantOut: "",
		// 	wantErr: true,
		// },
		{
			name: "two commands",
			cmds: []Command{
				{
					Executable: "echo",
					Args:       []string{"-n", "hello"},
				},
				{Executable: "cat"},
			},
			wantFunc: func(t *testing.T, got string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "hello", got)
			},
		},
		{
			name: "multiple commands",
			cmds: []Command{
				{
					Executable: "echo",
					Args:       []string{"-n", "hello"},
				},
				{Executable: "cat"}, {Executable: "cat"}, {Executable: "cat"},
			},
			wantFunc: func(t *testing.T, got string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "hello", got)
			},
		},
		// {
		// 	name: "last failed",
		// 		cmds: []Command{
		// 			{
		// 				Executable: "echo",
		// 				Args:       []string{"first"},
		// 			},
		// 			{Executable: "cat"}, {Executable: "cat"}, {Executable: "sh", Args: []string{"-c", ">&2 echo \"error\";exit 1"}},
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "middle failed",
		// 		cmds: []Command{
		// 			{
		// 				Executable: "echo",
		// 				Args:       []string{"first"},
		// 			},
		// 			{Executable: "cat"}, {Executable: "sh", Args: []string{"-c", ">&2 echo \"error\";exit 1"}}, {Executable: "cat"},
		// 	},
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := NewExecutor().RunR(tt.cmds...)
			tt.wantFunc(t, readOut(t, out), err)
			defer func(out io.ReadCloser) {
				if out != nil {
					_ = out.Close()
				}
			}(out)
		})
	}
}

func readOut(t *testing.T, out io.ReadCloser) string {
	if out == nil {
		return ""
	}
	got, err := io.ReadAll(out)
	if err != nil {
		t.Errorf("Read out failed. error=%s", err)
	}
	return string(got)
}
