package kill

import "syscall"

// Term はプロセスにSIGTERMを送る。
func Term(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

// Force はプロセスにSIGKILLを送る。
func Force(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}
