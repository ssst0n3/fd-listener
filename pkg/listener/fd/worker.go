package fd

import (
	"fmt"
	"github.com/ssst0n3/fd-listener/pkg"
	"os"
	"sort"
	"strconv"
	"sync"
)

type Worker struct {
	pid   int
	Stop  chan bool
	store sync.Map
	max   int
}

func NewWorker(pid int) (w *Worker) {
	w = &Worker{
		pid:  pid,
		Stop: make(chan bool),
	}
	go w.Work()
	return
}

func (l *Worker) Work() {
	for {
		select {
		case <-l.Stop:
			return
		default:
			l.do()
		}
	}
}

func (l *Worker) stat(fd int) (stat Stat, err error) {
	fdPath := fmt.Sprintf("/proc/%d/fd/%d", l.pid, fd)
	realPath, _ := os.Readlink(fdPath)
	if realPath == "" {
		realPath = "?"
	}
	stat = Stat{
		FdPath:   fdPath,
		RealPath: realPath,
	}
	socketPath, err := Socket(l.pid, realPath)
	if err != nil {
		return
	}
	stat.SocketPath = socketPath
	leak, err := Leak(fdPath, realPath)
	if err != nil {
		return
	}
	stat.Leak = leak
	flags, err := pkg.ReadFlags(fmt.Sprintf("/proc/%d/fdinfo/%d", l.pid, fd))
	if err != nil {
		return
	}
	stat.Flags = flags
	return
}

func (l *Worker) do() {
	_, err := os.Lstat(fmt.Sprintf("/proc/%d/", l.pid))
	if os.IsNotExist(err) {
		return
	}
	entries, err := os.ReadDir(fmt.Sprintf("/proc/%d/fd", l.pid))
	if err != nil {
		fmt.Printf("open /proc/%d/fd failed\n", l.pid)
		return
	}
	var fds []int
	for _, fd := range entries {
		fd, err := strconv.Atoi(fd.Name())
		if err != nil {
			continue
		}
		fds = append(fds, fd)
	}
	sort.Ints(fds)

	var changed bool

	// clear empty fd
	if len(fds) > 0 {
		for i := fds[len(fds)-1] + 1; i < l.max; i++ {
			if _, ok := l.store.LoadAndDelete(i); ok {
				changed = true
			}
		}
		l.max = fds[len(fds)-1]
	}

	var last int
	for _, fd := range fds {
		for i := last + 1; i < fd; i++ {
			if _, ok := l.store.LoadAndDelete(i); ok {
				changed = true
			}
		}
		last = fd
		stat, _ := l.stat(fd)
		if old, ok := l.store.Load(fd); !ok {
			l.store.Store(fd, stat)
			changed = true
		} else {
			if old != stat {
				l.store.Store(fd, stat)
				changed = true
			}
		}
	}
	if changed {
		l.print()
	}
}

func (l *Worker) print() {
	var keys []int
	l.store.Range(func(key any, value any) bool {
		fd := key.(int)
		keys = append(keys, fd)
		return true
	})
	sort.Ints(keys)
	for _, fd := range keys {
		stat, _ := l.store.Load(fd)
		fmt.Println(stat)
	}
	fmt.Println("----------------")
}