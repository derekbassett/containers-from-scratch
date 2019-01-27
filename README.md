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
  - Eli Uriegas How Docker Images Work: Union File Systems For Dummies https://www.terriblecode.com/blog/how-docker-images-work-union-file-systems-for-dummies/

## Demo:

##### Step 0. Bring the Vagrant Box up and Accessing the Ubuntu Host VM

```
$ vagrant up
```
This will take a while and will:
* Install the latest version of Go
* Untar root filesystems for ubuntu

If it fails due to internet connection re-run with

```
$ vagrant up --provision
```

If the box becomes outdated run

```
$ vagrant box update
```

Worse case scenario sometimes the box can get into a wonky state if it has been weeks since you
 have run the demo.  If that happens, delete the box and rerun `vagrant up`.  This is also the
likely fix if source files in this directory do not appear under `/src` in the vagrant box. 

Need to be running as Root

```shell
$ vagrant ssh
vagrant@ubuntu-xenial:~$ sudo -i
root@ubuntu-xenial:~# cd /src
```

##### Step 1. First example with No isolation

We are going to start with a very basic application.  Takes in a set of command line arguments and executes a command.

[embedmd]:# (step_1_no_isolation.go)
```go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

// go run step_1_no_isolation.go run <cmd> <args>
func main() {
	if len(os.Args) < 2 {
		panic("You must have at least two command line arguments")
	}
	switch os.Args[1] {
	case "run":
		run()
	default:
		panic("unknown command")
	}
}

func run() {
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

```shell
root@ubuntu-xenial:/src# go run main.go run echo "Hello World"
Running [echo Hello World]
Hello World
```

#### Step 2. Hostname isolation

##### Break it: Run bash with no isolation from inside a container

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

##### Fix it: Using Unix Timeshare System Isolation

Step 1: Add the following to main.go a `syscall` in the imports, and the following after `cmd.Stderr = os.Stderr`

[embedmd]:# (step_2_hostname_isolation.go /cmd.SysProcAttr/ /\}/)
```go
cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}
```

And the following to the import section

[embedmd]:# (step_2_hostname_isolation.go /import/ /\)/)
```go
import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)
```

`syscall.CLONE_NEWUTS` stands for clone new Unix Time Sharing System and the import section is required to access the syscall
package.

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

#### Step 3. Process isolation
If you can't get the last step to work you can start with `step_2_hostname_isolation.go`

##### Break it: Run ps with no isolation from inside a container

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

##### Fix it Part One: Re-running with PID isolation

Add the following to the main.go inside the `SysProcAttr`

[embedmd]:# (step_3_process_isolation.go /cmd.SysProcAttr/ /\}/)
```go
cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
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

##### Fix it Part Two: Re-run with Fork/Exec

Step 1: Duplicate the `run` function and create a `child` function.

Step 2: The run command `exec.Command` function executes `/proc/self/exe` creates a slice starting with the string `"child"`, and then copies
all the arguments from the second command line argument on, that is what the inner `...` is for, into the `append` function
which adds it to the slice.  The outer `...` unwinds the slice as variable arguments into Command.

Step 3:. Remove `cmd.SysProcAttr` in the `child` function, since the namespace has already been setup.

Step 4:. Now modify `main` to call `child` if first argument is `child`

Step 5: Set the container name using `must(syscall.Sethostname([]byte("container")))`


The functions are now:

[embedmd]:# (step_3_process_isolation.go /func main/ $)
```go
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

	must(syscall.Sethostname([]byte("container")))
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
```

Run the command again

```shell
root@ubuntu-xenial:/src# go run main.go run /bin/bash
running [/bin/bash] as PID 1
root@container:/src#
```

The command now is now correctly running as PID 1

#### Step 4. Having PS correctly work
If you can't get the last step to work you can start with `step_3_process_isolation.go`

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

##### Fix it: Generate an isolated filesystem

Included in the Vagrant file system with another root filesystem.  

Step 1: Add the following Commands into `child`

Step 2: Set the root directory using `must(syscall.Chroot("/rootfs-ubuntu"))`

Step 3: Set the current working directory using `must(os.Chdir("/"))`

Step 4: Mount the directory `/proc` using `must(syscall.Mount("proc", "proc", "proc", 0, ""))`

Step 5: Unmount the director `/proc` after the command is completed `must(syscall.Unmount("proc", 0))`


[embedmd]:# (step_4_fix_ps.go /func child/ /\}/)
```go
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
```

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
I_AM_A_CONTAINER_ROOT_FS  bin  boot  dev  etc  home  lib  lib64  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
root@container:/# exit
exit
root@ubuntu-xenial:/src#
```

**Potential exploit**
Check out this potential exploit at this Gist https://gist.github.com/derekbassett/c67c0b129804c55ec3ce2cbdf1412985

#### Step 5. Overlay File system
If you can't get the last step to work you can start with `step_4_fix_ps.go`

This is all great but we have a problem, if we change the files in container we change those files for every container instance 
using those files.  The simple solution is to use overlay(fs) to make the ubuntu root file system read-only, and make a new
modifiable root file system. 

```shell
root@ubuntu-xenial:/src# go run main.go run /bin/bash
running [/bin/bash] as PID 1
root@container:/# ls
I_AM_CONTAINER_ROOT_FS  bin  boot  dev  etc  home  lib  lib64  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
root@container:/# touch NEW_FILE
root@container:/# ls
I_AM_CONTAINER_ROOT_FS  NEW_FILE  bin  boot  dev  etc  home  lib  lib64  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
root@container:/# exit
exit
root@ubuntu-xenial:/src# ls /rootfs-ubuntu
bin  boot  dev  etc  home  I_AM_CONTAINER_ROOT_FS  lib  lib64  media  mnt  NEW_FILE  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
```

Use an overlay file system

```shell
root@ubuntu-xenial:/src# go run main.go run /bin/bash
running [/bin/bash] as PID 1
root@container:/# mkdir base diff overlay workdir
root@container:/# touch base/MY_BASE_FILE
root@container:/# mount -t overlay -o lowerdir=base,upperdir=diff,workdir=workdir overlay overlay
```

We now have an overlay directory.  If you write into 

##### Fix It: Make an overlay file system

# TODO: Figure out how to do this.

#### Step 6. CGroup

Namespaces are what you can see. CGroups what you can use, and allow you to control the resources.  

Follow along with the group:

```script
root@ubuntu-xenial:/src# apt install docker.io
root@ubuntu-xenial:/src# docker run -it ubuntu /bin/bash
root@9656e6202efd:/# exit
root@ubuntu-xenial:/src# cd /sys/fs/cgroup/memory
root@ubuntu-xenial:/sys/fs/cgroup/memory# ls
cgroup.clone_children  memory.failcnt                  memory.kmem.tcp.failcnt             memory.max_usage_in_bytes        memory.stat            system.slice
cgroup.event_control   memory.force_empty              memory.kmem.tcp.limit_in_bytes      memory.move_charge_at_immigrate  memory.swappiness      tasks
cgroup.procs           memory.kmem.failcnt             memory.kmem.tcp.max_usage_in_bytes  memory.numa_stat                 memory.usage_in_bytes  user.slice
cgroup.sane_behavior   memory.kmem.limit_in_bytes      memory.kmem.tcp.usage_in_bytes      memory.oom_control               memory.use_hierarchy
docker                 memory.kmem.max_usage_in_bytes  memory.kmem.usage_in_bytes          memory.pressure_level            notify_on_release
init.scope             memory.kmem.slabinfo            memory.limit_in_bytes               memory.soft_limit_in_bytes       release_agent
root@ubuntu-xenial:/sys/fs/cgroup/memory# ls docker
cgroup.clone_children  memory.kmem.failcnt             memory.kmem.tcp.limit_in_bytes      memory.max_usage_in_bytes        memory.soft_limit_in_bytes  notify_on_release
cgroup.event_control   memory.kmem.limit_in_bytes      memory.kmem.tcp.max_usage_in_bytes  memory.move_charge_at_immigrate  memory.stat                 tasks
cgroup.procs           memory.kmem.max_usage_in_bytes  memory.kmem.tcp.usage_in_bytes      memory.numa_stat                 memory.swappiness
memory.failcnt         memory.kmem.slabinfo            memory.kmem.usage_in_bytes          memory.oom_control               memory.usage_in_bytes
memory.force_empty     memory.kmem.tcp.failcnt         memory.limit_in_bytes               memory.pressure_level            memory.use_hierarchy
root@ubuntu-xenial:/sys/fs/cgroup/memory#
root@ubuntu-xenial:/sys/fs/cgroup/memory# cat docker/cgroup.procs
root@ubuntu-xenial:/sys/fs/cgroup/memory#

```

Step 1: Transfer this code into main.go
```
func cg(name string) {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.MkdirAll(filepath.Join(pids, name), 0755)
	must(ioutil.WriteFile(filepath.Join(pids, name, "pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, name, "notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, name, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}
```
Step 2: Add the following line to `child()`
```
    cg("scratch")
```
Step 3: Add the following import
```
    "io/ioutil"
    "path/filepath"
    "strconv"

```

```go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// go run main.go run <cmd> <args>
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

	cg("scratch")

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

func cg(name string) {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.MkdirAll(filepath.Join(pids, name), 0755)
	must(ioutil.WriteFile(filepath.Join(pids, name, "pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, name, "notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, name, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
```

##### Fork Bomb
**WARNING THIS CAN IF PUT IN THE WRONG TERMINAL CAUSE YOU TO HAVE TO REBOOT YOUR MACHINE**

Inside the container bash area type in the following command  
```shell
root@ubuntu-xenial:/src# go run main.go run /bin/bash
running [/bin/bash] as PID 1
root@container:/# :() { : | : & }; :
```
In an alternate container run `ps aux`

## Installation

You must have [Git](https://gist.github.com/derhuerst/1b15ff4652a867391f03), [Vagrant](https://www.vagrantup.com/downloads.html) and
[Virtual Box](https://www.virtualbox.org/wiki/Downloads) installed.
