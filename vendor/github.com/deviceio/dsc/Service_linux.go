package dsc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"text/template"
)

var systemvTemplate = `
#!/bin/sh
### BEGIN INIT INFO
# Provides:
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start daemon at boot time
# Description:       Enable service provided by daemon.
### END INIT INFO

dir=""
cmd="{{.cmd}}"
user=""

name=` + "`basename $0`" + `
pid_file="/var/run/$name.pid"
stdout_log="/var/log/$name.log"
stderr_log="/var/log/$name.err"

get_pid() {
    cat "$pid_file"
}

is_running() {
    [ -f "$pid_file" ] && ps ` + "`get_pid`" + ` > /dev/null 2>&1
}

case "$1" in
    start)
    if is_running; then
        echo "Already started"
    else
        echo "Starting $name"
        cd "$dir"
        if [ -z "$user" ]; then
            sudo $cmd >> "$stdout_log" 2>> "$stderr_log" &
        else
            sudo -u "$user" $cmd >> "$stdout_log" 2>> "$stderr_log" &
        fi
        echo $! > "$pid_file"
        if ! is_running; then
            echo "Unable to start, see $stdout_log and $stderr_log"
            exit 1
        fi
    fi
    ;;
    stop)
    if is_running; then
        echo -n "Stopping $name.."
        kill ` + "`get_pid`" + `
        for i in {1..10}
        do
            if ! is_running; then
                break
            fi

            echo -n "."
            sleep 1
        done
        echo

        if is_running; then
            echo "Not stopped; may still be shutting down or shutdown may have failed"
            exit 1
        else
            echo "Stopped"
            if [ -f "$pid_file" ]; then
                rm "$pid_file"
            fi
        fi
    else
        echo "Not running"
    fi
    ;;
    restart)
    $0 stop
    if is_running; then
        echo "Unable to stop, will not attempt to start"
        exit 1
    fi
    $0 start
    ;;
    status)
    if is_running; then
        echo "Running"
    else
        echo "Stopped"
        exit 1
    fi
    ;;
    *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac

exit 0
`

func (t *Service) Apply() (bool, error) {
	var absent bool
	var args []string
	var changed = false
	var err error
	var name string
	var path string
	var started bool
	var initFile string

	if name, err = t.name(); err != nil {
		return false, err
	}

	if path, err = t.path(); err != nil {
		return false, err
	}

	if started, err = t.started(); err != nil {
		return false, err
	}

	if args, err = t.args(); err != nil {
		return false, err
	}

	initFile = fmt.Sprintf("/etc/init.d/%v", name)

	_, err = os.Stat(initFile)

	if os.IsNotExist(err) && absent {
		return false, nil
	}

	if os.IsNotExist(err) && !absent {
		changed = true

		if err = ioutil.WriteFile(initFile, []byte(systemvTemplate), 0600); err != nil {
			return changed, err
		}
	}

	if err == nil && absent {
		if err = os.Remove(initFile); err != nil {
			return false, err
		}
		return true, nil
	}

	content, err := ioutil.ReadFile(initFile)

	if err != nil {
		return changed, err
	}

	var initTemplate *bytes.Buffer
	templ, err := template.New("").Parse(systemvTemplate)

	if err != nil {
		return false, err
	}

	pathAndArgs := strings.Join(
		append([]string{path}, args...),
		" ",
	)

	if err = templ.Execute(initTemplate, struct {
		cmd string
	}{
		cmd: pathAndArgs,
	}); err != nil {
		return changed, err
	}

	if !reflect.DeepEqual(content, initTemplate.Bytes()) {
		changed = true

		if err = ioutil.WriteFile(initFile, initTemplate.Bytes(), 0600); err != nil {
			return changed, err
		}
	}

	cmdout, err := exec.Command(initFile, "status").Output()

	if err != nil {
		return changed, err
	}

	if string(cmdout) == "Running" && !started {
		changed = true

		if _, err := exec.Command(initFile, "stop").Output(); err != nil {
			return changed, err
		}
	}

	if string(cmdout) == "Stopped" && started {
		changed = true

		if _, err := exec.Command(initFile, "stop").Output(); err != nil {
			return changed, err
		}
	}

	return changed, nil
}

func (t *Service) Refresh() error {
	return fmt.Errorf("Not Yet Supported")
}

func (t *Service) Requires() []string {
	if t.Relation == nil || t.Relation.Require == nil {
		return []string{}
	}

	return t.Relation.Require
}

func (t *Service) Notifies() []string {
	if t.Relation == nil || t.Relation.Notify == nil {
		return []string{}
	}

	return t.Relation.Notify
}

func (t *Service) Refreshes() []string {
	if t.Relation == nil || t.Relation.Refresh == nil {
		return []string{}
	}

	return t.Relation.Refresh
}
