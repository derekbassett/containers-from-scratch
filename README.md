# Building container from scratch using Go

> Building containers from scratch using Go as seen from @lizrice presentation at DockerCon 2017
 
This code will not build on Mac or Windows it will only work with GOOS=Linux, hence why we include the Vagrant files.

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
		panic("You must have commnad line two arguments")
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


## Installation

You must have [Vagrant](https://www.vagrantup.com/downloads.html) and 
[Virtual Box](https://www.virtualbox.org/wiki/Downloads) installed.
