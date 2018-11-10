# Building container from scratch using Go

> Building containers from scratch using Go 
 
This code will not build on Mac or Windows instead it will only work with GOOS=Linux, hence why we include the Vagrant files.
See instructions at the bottom to get Vagrant configured for your system.

## Links:

  - Gist from Liz Rice: https://gist.github.com/lizrice/a5ef4d175fd0cd3491c7e8d716826d27
  - Gist from Julian Friedman: https://gist.github.com/julz/c0017fa7a40de0543001
  - Julian post at InfoQ: https://www.infoq.com/articles/build-a-container-golang
  - Liz Rice talk at Golang UK Conference 2016: https://www.youtube.com/watch?v=HPuvDm8IC-4
  - Liz Rice talk at Container Camp UK 2016: https://www.youtube.com/watch?v=Utf-A4rODH8
  - Wellington Silva Containers from Scratch Demo https://github.com/wsilva/container-from-scratch-demo

## Demo:

##### 1. Bring the Vagrant Box up

```
$ vagrant up
```
This will take a while and will:
* Install the latest version of Go
* Untar root filesystems for alpine, centos, debian, fedora, ubuntu
 
If it fails due to internet connection re-run with

```
$ vagrant up --provision
```

If the box becomes outdated run

```
$ vagrant box update
```

##### 2. Accessing the Ubuntu Host VM

```shell
$ vagrant ssh
vagrant@ubuntu-xenial:~$ sudo -i
root@ubuntu-xenial:~# cd /src
```

##### 3. First example with No isolation

```shell
root@ubuntu-xenial:/src# go run main.go run echo "Hello World"
Running [echo Hello World]
Hello World
```

##### 4. Run bash with no isolation from inside a container

```shell
root@ubuntu-xenial:/src# go run main.go run /bin/bash
Running [/bin/bash]
root@ubuntu-xenial:/src# hostname
ubuntu-xenial
root@ubuntu-xenial:/src# hostname demo
root@ubuntu-xenial:/src# hostname
demo
root@ubuntu-xenial:/src# exit
exit
root@ubuntu-xenial:/src# hostname
demo
root@ubuntu-xenial:/src#
```

The hostname was changed in the host.

Before going to the next step reset the hostname

```shell
root@ubuntu-xenial:/src# hostname ubuntu-xenial
```

##### 5. Re-running with UTS isolation

Add the following to main.go a `syscall` in the imports, and the following after `cmd.Stderr = os.Stderr`
```
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS,
    } 
```

`syscall.CLONE_NEWUTS` stands for clone new Unix Time Sharing System.

```shell
root@ubuntu-xenial:/src# go run main.go run /bin/bash
Running [/bin/bash]
root@ubuntu-xenial:/src# hostname
ubuntu-xenial
root@ubuntu-xenial:/src# hostname demo
root@ubuntu-xenial:/src# hostname
demo
root@ubuntu-xenial:/src# exit
exit
root@ubuntu-xenial:/src# hostname
ubuntu-xenial
root@ubuntu-xenial:/src#
```

The hostname is not modified inside the container.

##### 6. Run ps with no isolation from inside a container

```shell
root@ubuntu-xenial:/src# ps
  PID TTY          TIME CMD
 2730 pts/0    00:00:00 sudo
 2731 pts/0    00:00:00 bash
 2852 pts/0    00:00:00 ps
root@ubuntu-xenial:/src# go run main.go run /bin/bash
Running [/bin/bash]
root@ubuntu-xenial:/src# ps
  PID TTY          TIME CMD
 2730 pts/0    00:00:00 sudo
 2731 pts/0    00:00:00 bash
 2853 pts/0    00:00:00 go
 2886 pts/0    00:00:00 main
 2890 pts/0    00:00:00 bash
 2900 pts/0    00:00:00 ps
root@ubuntu-xenial:/src# exit
exit
root@ubuntu-xenial:/src#
```

The list of processes include the parent process

##### 7. Re-running with PID isolation

Add the following to the main.go inside the `SysProcAttr`
```
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags : syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
    }
```

```shell
root@ubuntu-xenial:/src# ps
  PID TTY          TIME CMD
 2730 pts/0    00:00:00 sudo
 2731 pts/0    00:00:00 bash
 2852 pts/0    00:00:00 ps
root@ubuntu-xenial:/src# go run main.go run /bin/bash
Running [/bin/bash]
root@ubuntu-xenial:/src# ps
  PID TTY          TIME CMD
 2730 pts/0    00:00:00 sudo
 2731 pts/0    00:00:00 bash
 2853 pts/0    00:00:00 go
 2886 pts/0    00:00:00 main
 2890 pts/0    00:00:00 bash
 2900 pts/0    00:00:00 ps
root@ubuntu-xenial:/src# exit
exit
root@ubuntu-xenial:/src#
```

No luck, but the reason why this isn't working is because in order to change the pid we need to fork/exec

##### 8. Re-run with Fork/Exec

Modify the program `main.go`
```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		panic("You must have at least two commnad line arguments")
	}
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("what???")
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	// Setup Stdin, Stdout, Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}
	must(cmd.Run())
}

func child() {
	fmt.Printf("running %v as PID %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	// Setup Stdin, Stdout, Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
```

1. Duplicate the `run` function and create a `child` function.

1. The run command `exec.Command` function executes `/proc/self/exe` creates a slice starting with the string `"child"`, and then copies
all the arguments from the second command line argument on, that is what the inner `...` is for, into the `append` function 
which adds it to the slice.  The outer `...` unwinds the slice as variable arguments into Command.

1. Remove `cmd.SysProcAttr` in the `child` function.

1. Now modify `main` to call `child` if first argument is `child`

Run the command again

```shell
root@ubuntu-xenial:/src# go run main.go run /bin/bash
running [/bin/bash] as PID 1
root@ubuntu-xenial:/src#
```

But if we run `ps` we still have the same list of processes

```
root@ubuntu-xenial:/src# ps
  PID TTY          TIME CMD
 1930 pts/0    00:00:00 sudo
 1931 pts/0    00:00:00 bash
 2116 pts/0    00:00:00 go
 2149 pts/0    00:00:00 main
 2153 pts/0    00:00:00 exe
 2157 pts/0    00:00:00 bash
 2167 pts/0    00:00:00 ps
root@ubuntu-xenial:/src# exit
exit
root@ubuntu-xenial:/src#
```

`ps` looks in `/proc` instead of at the process itself.  In order for `ps` to correctly work we need to generate
it's own file system.

##### 9. Generate an isolated filesystem

Included in the Vagrant file system are five root filesystems.  Alpine, Centos, Debian, Fedora, Ubuntu.  This example will
use Ubuntu but you can select the version you want based on rootfs.

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		panic("You must have at least two commnad line arguments")
	}
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("what???")
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	// Setup Stdin, Stdout, Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	must(cmd.Run())
}

func child() {
	fmt.Printf("running %v as PID %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	// Setup Stdin, Stdout, Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/rootfs-ubuntu"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	must(cmd.Run())
	must(syscall.Unmount("proc", 0))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
```

1. Add a new flag to `Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,` in SysProcAttr

1. Add a new flag `Unshareflags: syscall.CLONE_NEWNS,` in SysProcAttr

1. Add the following Commands into `child`
    1. Set the container name using `must(syscall.Sethostname([]byte("container")))`
    1. Set the root directory using `must(syscall.Chroot("/rootfs-ubuntu"))`
    1. Set the current working directory using `must(os.Chdir("/"))`
    1. Mount the directory `/proc` using `must(syscall.Mount("proc", "proc", "proc", 0, ""))`
    1. Unmount the director `/proc` after the command is completed `must(syscall.Unmount("proc", 0))`
    

```shell
root@ubuntu-xenial:/src# ls /
bin   etc                   initrd.img  lost+found  opt   rootfs-alpine  rootfs-fedora  sbin  srv  usr
boot  home                  lib         media       proc  rootfs-centos  rootfs-ubuntu  snap  sys  var
dev   I_AM_THE_HOST_ROOTFS  lib64       mnt         root  rootfs-debian  run            src   tmp  vmlinuz
root@ubuntu-xenial:/src# go run main.go run /bin/bash
running [/bin/bash] as PID 1
root@container:/# ps
  PID TTY          TIME CMD
    1 ?        00:00:00 exe
    5 ?        00:00:00 bash
   10 ?        00:00:00 ps
root@container:/# ls
I_AM_A_ROOT_FS  bin  boot  dev  etc  home  lib  lib64  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
root@container:/# exit
exit
root@ubuntu-xenial:/src#
```

##### 10. Adding CGroup controls

CGroups allow control of the contained process.  We are going to add some simple cgroup functionality to constrain 

```go

```

## Installation

You must have [Vagrant](https://www.vagrantup.com/downloads.html) and 
[Virtual Box](https://www.virtualbox.org/wiki/Downloads) installed.
