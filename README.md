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

1. Bring the Vagrant Box up
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

2. Accessing the Ubuntu Host VM

```shell
$ vagrant ssh
vagrant@ubuntu-xenial:~$ sudo -i
root@ubuntu-xenial:~# cd /src
```

3. First example with No isolation

```shell
root@ubuntu-xenial:/src# go run main.go run echo "Hello World"
Running [echo Hello World]
Hello World
```

4. Run bash with no isolation from inside a container

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

5. Re-running with UTS isolation

Add the following to main.go

```go

```


## Installation

You must have [Vagrant](https://www.vagrantup.com/downloads.html) and 
[Virtual Box](https://www.virtualbox.org/wiki/Downloads) installed.
