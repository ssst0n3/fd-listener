package stat

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ssst0n3/fd-listener/pkg"
	"os"
)

type Stat struct {
	FdPath     string
	RealPath   string
	SocketPath string
	Leak       bool
	Flags      int64
	Changed    bool
}

func New(pid, fd int) (stat *Stat, err error) {
	fdPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	realPath, _ := os.Readlink(fdPath)
	if realPath == "" {
		realPath = "?"
	}
	stat = &Stat{
		FdPath:   fdPath,
		RealPath: realPath,
	}
	socketPath, err := Socket(pid, realPath)
	if err != nil {
		return
	}
	stat.SocketPath = socketPath
	leak, err := Leak(fdPath, realPath)
	if err != nil {
		return
	}
	stat.Leak = leak
	flags, err := pkg.ReadFlags(fmt.Sprintf("/proc/%d/fdinfo/%d", pid, fd))
	if err != nil {
		return
	}
	stat.Flags = flags
	return
}

func (s *Stat) Change(c bool) {
	s.Changed = c
}

func (s *Stat) String() (content string) {
	fdPath := s.FdPath
	if s.Changed {
		fdPath = color.New(color.Underline).Sprint(fdPath)
	}
	leaked := ""
	if s.Leak {
		leaked = color.RedString("leaked!")
	}
	flags := pkg.ParseFlags(s.Flags)
	var socketPath string
	if s.SocketPath != "" {
		socketPath = " -> " + s.SocketPath
	}
	content = fmt.Sprintf("%s -> %s%s\t; %s\t%s", fdPath, s.RealPath, socketPath, leaked, flags)
	//content = fmt.Sprintf("%-30s -> %-30s%-15s; %-10s %s", fdPath, s.RealPath, socketPath, leaked, flags)
	return
}

func (s *Stat) Equals(s2 *Stat) bool {
	return s.FdPath == s2.FdPath &&
		s.RealPath == s2.RealPath &&
		s.SocketPath == s2.SocketPath &&
		s.Leak == s2.Leak &&
		s.Flags == s2.Flags
}
