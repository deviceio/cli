package dsc

import (
	"os/exec"
)

// Exec represents a local execution of a command
type Exec struct {
	*Relation

	// Cmd the binary to execute.
	Cmd string

	// CmdFunc the function to call to obtain the binary to execute. If CmdFunc
	// is non nil, Cmd is ignored. Function may return an optional error to halt
	// enforcement.
	CmdFunc func(*Exec) (string, error)

	// Args the argument to supply to the executing Cmd
	Args []string

	// ArgsFunc the function to call to obtain the Cmd arguments. If ArgsFunc is
	// non nil, Args is ignored. Function may return an optional error to halt
	// enforcement
	ArgsFunc func(*Exec) ([]string, error)
}

// Apply ...
func (t *Exec) Apply() (bool, error) {
	cmd, err := t.cmd()

	if err != nil {
		return false, err
	}

	args, err := t.args()

	if err != nil {
		return false, err
	}

	e := exec.Command(cmd, args...)

	err = e.Run()

	if err != nil {
		return false, err
	}

	return true, nil
}

func (t *Exec) Refresh() error {
	return nil
}

func (t *Exec) Requires() []string {
	if t.Relation == nil || t.Relation.Require == nil {
		return []string{}
	}

	return t.Relation.Require
}

func (t *Exec) Notifies() []string {
	if t.Relation == nil || t.Relation.Notify == nil {
		return []string{}
	}

	return t.Relation.Notify
}

func (t *Exec) Refreshes() []string {
	if t.Relation == nil || t.Relation.Refresh == nil {
		return []string{}
	}

	return t.Relation.Refresh
}

func (t *Exec) args() ([]string, error) {
	if t.ArgsFunc != nil {
		return t.ArgsFunc(t)
	}

	if t.Args == nil {
		return []string{}, nil
	}

	return t.Args, nil
}

func (t *Exec) cmd() (string, error) {
	if t.CmdFunc != nil {
		return t.CmdFunc(t)
	}

	return t.Cmd, nil
}
