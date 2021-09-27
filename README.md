# mk

_Multiplatform make-like tool_

Ok, ok - you say - why another make tool clone? Oh lord, there we go again....

Indeed its a valid question... reasons are:

1 - Cuz I could not find a make tool that fit my needs as simple as this one

2 - Cuz windows is always a pain when it comes to compiling software

3 - Why not?

4 - Did I mention I hate verbose stuff?

So these are the drivers for writting mk

## Install

```shell
go install github.com/digitalcircle-com-br/mk@latest
```

## How it works:

mk will look for mk files (which may be named: mk, mk.yaml .mk or .mk.yaml)

you may create a new mk file by using ```mk -i``` bingo, thats all...

The file will look like this one:

```yaml
# SAMPLE mk file - Feel free to add your own header
default: a #This is the default task, in case you call command w/o parameters
env: # in case you want to add var to env, you may add it here
  a: 1
  b: 2

tasks: #now lets define the tasks 
  a_darwin_arm64: #this is the task name
    cmd: |- # and this is the command - which may be multiline, no issues.
      echo \"${TASK} / ${BASETASK}\"
      ls -larth
      pwd
  deploy:
    pre: [ build,test ] # pre is an array of predecessors
    help: Deploys the project # help prints the help message
    cmd: echo deploying
  test:
    pre: [ c ]
    help: Tests project
    cmd: echo testing
  build:
    help: Build binaries
    cmd: echo building
  main:
    help: Main task
    pre: [ build ]
    cmd: |-
      echo main
      echo done%
```

And thats it.

## Some gotchas you should notice

### Name resolution

Tasks are resolved considering this rule: task_os_arch: In case you have a task with the os name and arch name, it will
have higher precedence at resolving it. Suppose you add 2 tasks in your mk file: a_windows_amd64 and a_darwin_arm64. In
case youre on a Mac with Apple silicon, and call make a, a_darwin_arm64 will be called. In case you have a task a_darwin
and a_windows, and call mk a from a Mac with Intel processor, it will call a_darwing. Lastly, in case you also define a
task a, it will be called in case none of these more restrictive rules find math.

> By adopting this approach same mk file will allow multiple platform compilation.

### Variables

You may place ${VAR} anywhere in your command, and it will be replaced by mk. It provides you some var, and also env
vars

## Help reference

```shell
Usage: mk [<tasks> ...]

Arguments:
  [<tasks> ...]    Tasks to be run - Default is main.

Flags:
  -h, --help              Show context-sensitive help.
  -f, --file=STRING       File to be used - Defaults are: .mk.yaml, .mk, mk, mk.yaml
  -i, --init              Creates a new empty file (default is .mk.yaml in case no filename is provided)
  -v, --ver               Prints version and exit
  -l, --list              Check file and print tasks
  -d, --dbg               Debugs execution
      --dump-validator    Dumps Validator JSON File
  -e, --env               Dumps env and vars

```

# TODO

- Output is pretty ugly, but the best I could think of so far...
- Integrate zip and git into "internal tasks"
- allow file to have tasks as strings in case no other props are required
- allow file include
- allow it to run as server
- Accepting recommendations on how to improve it