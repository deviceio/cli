package dsc

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
)

// File represents a normal file to enforce on the local filesystem.
type File struct {
	*Relation

	// Path specifies the path to the file you wish to enforce
	Path string

	// PathFunc provides a function that returns a path for this file resource.
	// if PathFunc is not nil, Path is ignored. Function may return an error to
	// halt enforcement.
	PathFunc func(*File) (string, error)

	// Absent indicates if this file should be Absent on the filesystem. Default
	// is false indicating that the file should be present. If Absent is true,
	// all other properties of enforcement are ignored (as expected)
	Absent bool

	// AbsentFunc MAY be provided to return a bool indicating if the file should
	// be Absent. If AbsentFunc is non nil, File.Absent is ignored. Function may
	// return an error to halt enforcement.
	AbsentFunc func(*File) (bool, error)

	// Content MAY be provided to enforce file content. If Content is nil no content
	// eforcement is done.
	Content []byte

	// ContentFunc MAY be provided a function which returns a []byte to enforce
	// file content. If ContentFunc is non nil, File.Content is ignored. Function
	// may return an error to halt enforcement.
	ContentFunc func(*File) ([]byte, error)

	// Mode MAY provide an os.FileMode to enforce on the resource. If Mode is 0
	// no Mode will be enforced.
	Mode os.FileMode

	// ModeFunc MAY be provide a function that returns a os.FileMode. If ModeFunc
	// is non nil, File.Mode is ignored. Function may return an error to
	// halt enforcement.
	ModeFunc func(*File) (os.FileMode, error)
}

// Apply enforces this file resource
func (t *File) Apply() (bool, error) {
	changed := false

	fpath, err := t.path()

	if err != nil {
		return false, err
	}

	absent, err := t.absent()

	if err != nil {
		return false, err
	}

	stat, err := os.Stat(fpath)

	if os.IsNotExist(err) && absent {
		return false, nil
	}

	if os.IsNotExist(err) && !absent {
		if err := os.MkdirAll(path.Dir(fpath), 0777); err != nil {
			return false, err
		}

		if err := ioutil.WriteFile(fpath, []byte(""), 0777); err != nil {
			return false, err
		}

		stat, err = os.Stat(fpath)

		if err != nil {
			return false, err
		}
	}

	if err == nil && absent {
		if err = os.Remove(fpath); err != nil {
			return false, err
		}
		return true, nil
	}

	if err != nil {
		return false, err
	}

	mode, err := t.mode()

	if err != nil {
		return false, err
	}

	if mode > 0 {
		if mode != stat.Mode() {
			if err := os.Chmod(fpath, mode); err != nil {
				return false, err
			}
			changed = true
		}
	}

	content, err := t.content()

	if err != nil {
		return false, err
	}

	if content == nil {
		return changed, nil
	}

	fcontent, err := ioutil.ReadFile(fpath)

	if err != nil {
		return false, err
	}

	if !reflect.DeepEqual(content, fcontent) {
		if err = ioutil.WriteFile(fpath, content, mode); err != nil {
			return false, err
		}
	}

	return changed, nil
}

func (t *File) Refresh() error {
	return nil
}

func (t *File) Requires() []string {
	if t.Relation == nil || t.Relation.Require == nil {
		return []string{}
	}

	return t.Relation.Require
}

func (t *File) Notifies() []string {
	if t.Relation == nil || t.Relation.Notify == nil {
		return []string{}
	}

	return t.Relation.Notify
}

func (t *File) Refreshes() []string {
	if t.Relation == nil || t.Relation.Refresh == nil {
		return []string{}
	}

	return t.Relation.Refresh
}

func (t *File) absent() (bool, error) {
	if t.AbsentFunc != nil {
		return t.AbsentFunc(t)
	}

	return t.Absent, nil
}

func (t *File) content() ([]byte, error) {
	if t.ContentFunc != nil {
		return t.ContentFunc(t)
	}

	return t.Content, nil
}

func (t *File) path() (string, error) {
	if t.PathFunc != nil {
		return t.PathFunc(t)
	}

	return t.Path, nil
}

func (t *File) mode() (os.FileMode, error) {
	if t.ModeFunc != nil {
		return t.ModeFunc(t)
	}

	return t.Mode, nil
}
