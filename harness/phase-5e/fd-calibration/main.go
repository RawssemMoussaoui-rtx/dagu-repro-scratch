package main

import "fmt"

const (
	fDupFD        = 0
	fDupFDCloexec = 1030
	atFDCWD       = -100
)

type fdTable struct {
	valid []bool
	gen   []uint64
	next  uint64
}

func newFDTable(size int) *fdTable {
	t := &fdTable{
		valid: make([]bool, size),
		gen:   make([]uint64, size),
	}
	for i := 0; i < 3 && i < size; i++ {
		t.next++
		t.valid[i] = true
		t.gen[i] = t.next
	}
	return t
}

func (t *fdTable) allocGen() uint64 {
	t.next++
	return t.next
}

func (t *fdTable) firstFree(start int) int {
	for i := start; i < len(t.valid); i++ {
		if !t.valid[i] {
			return i
		}
	}
	return -1
}

func (t *fdTable) open() int {
	fd := t.firstFree(0)
	if fd < 0 {
		return -1
	}
	t.valid[fd] = true
	t.gen[fd] = t.allocGen()
	return fd
}

func (t *fdTable) openPath(path string) int {
	if path == "" {
		return -1
	}
	return t.open()
}

func (t *fdTable) openat(dirfd int, path string) int {
	if path == "" {
		return -1
	}
	if dirfd != atFDCWD {
		if dirfd < 0 || dirfd >= len(t.valid) || !t.valid[dirfd] {
			return -1
		}
	}
	return t.open()
}

func (t *fdTable) close(fd int) int {
	if fd < 0 || fd >= len(t.valid) || !t.valid[fd] {
		return -1
	}
	t.valid[fd] = false
	return 0
}

func (t *fdTable) dup(oldfd int) int {
	if oldfd < 0 || oldfd >= len(t.valid) || !t.valid[oldfd] {
		return -1
	}
	newfd := t.firstFree(0)
	if newfd < 0 {
		return -1
	}
	t.valid[newfd] = true
	t.gen[newfd] = t.allocGen()
	return newfd
}

func (t *fdTable) dup2(oldfd, newfd int) int {
	if oldfd < 0 || newfd < 0 || oldfd >= len(t.valid) || newfd >= len(t.valid) {
		return -1
	}
	if !t.valid[oldfd] {
		return -1
	}
	if newfd != oldfd && t.valid[newfd] {
		t.valid[newfd] = false
	}
	t.valid[newfd] = true
	t.gen[newfd] = t.allocGen()
	return newfd
}

func (t *fdTable) fcntlDupfd(fd, arg int, cloexec bool) int {
	_ = cloexec
	if fd < 0 || fd >= len(t.valid) || !t.valid[fd] || arg < 0 {
		return -1
	}
	newfd := t.firstFree(arg)
	if newfd < 0 {
		return -1
	}
	t.valid[newfd] = true
	t.gen[newfd] = t.allocGen()
	return newfd
}

func (t *fdTable) state(label string) {
	fmt.Printf("state %s:", label)
	for i := 0; i < len(t.valid); i++ {
		if t.valid[i] {
			fmt.Printf(" %d:v1g%d", i, t.gen[i])
		} else {
			fmt.Printf(" %d:v0", i)
		}
	}
	fmt.Println()
}

func main() {
	t := newFDTable(16)

	t.state("initial")
	fd := t.openPath("/tmp/alpha")
	fmt.Printf("open /tmp/alpha -> fd=%d gen=%d\n", fd, t.gen[fd])

	fd2 := t.openat(atFDCWD, "/tmp/beta")
	fmt.Printf("openat AT_FDCWD /tmp/beta -> fd=%d gen=%d\n", fd2, t.gen[fd2])

	fmt.Printf("open \"\" -> fd=%d\n", t.openPath(""))

	if err := t.close(fd); err == 0 {
		fmt.Printf("close %d -> 0\n", fd)
	}
	reused := t.openPath("/tmp/gamma")
	fmt.Printf("open /tmp/gamma -> fd=%d gen=%d (reuse)\n", reused, t.gen[reused])

	dupfd := t.dup(reused)
	fmt.Printf("dup %d -> fd=%d gen=%d\n", reused, dupfd, t.gen[dupfd])

	dup2fd := t.dup2(dupfd, 7)
	fmt.Printf("dup2 %d 7 -> fd=%d gen=%d\n", dupfd, dup2fd, t.gen[dup2fd])

	fdup := t.fcntlDupfd(reused, 4, false)
	fmt.Printf("fcntl F_DUPFD %d arg=4 -> fd=%d gen=%d\n", reused, fdup, t.gen[fdup])

	fdupce := t.fcntlDupfd(reused, 6, true)
	fmt.Printf("fcntl F_DUPFD_CLOEXEC %d arg=6 -> fd=%d gen=%d\n", reused, fdupce, t.gen[fdupce])

	alias := t.dup2(reused, fd2)
	fmt.Printf("dup2 %d %d -> fd=%d gen=%d (implicit close)\n", reused, fd2, alias, t.gen[alias])

	if err := t.close(0); err == 0 {
		fmt.Println("close 0 -> 0")
	}
	stdinReplacement := t.openPath("/tmp/stdin-replacement")
	fmt.Printf("open /tmp/stdin-replacement -> fd=%d gen=%d (fd 0 reuse)\n", stdinReplacement, t.gen[stdinReplacement])

	t.state("final")
}
