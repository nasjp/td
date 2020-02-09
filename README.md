# [WIP] td

Super Simple TODO management tool

## Usage

```text
NAME:
   td - super simple TODO management tool

USAGE:
   td command [command options] [arguments...]

COMMANDS:
     new,  n    create todo
     list, l    list todo
     add,  a    add new todo
     done, d    make todo status done
     open, o    make todo status open
```

## Installation

```sh
go get github.com/nasjp/td
```

## Example

```sh
$ td new
.td.json was generated

$ td add
task's detail:  This is a task
priority [H/L]:

$ td list
  1 [ ] This is a task

$ td add
task's detail:  This is a high priority task
priority [H/L]: H

$ td list
  1 [ ] This is a task
H 2 [ ] This is a high priority task

$ td done 2

$ td list
  1 [ ] This is a task

$ td list -a
  1 [ ] This is a task
H 2 [x] This is a high priority task
```

## TODO

- enable to use sub-tasks
- enable to use select current main-task

## License

MIT
