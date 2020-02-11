package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

const filename = ".td.json"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Error: %w\n", err))
		cmdHelp()
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	if len(os.Args) <= 1 {
		return fmt.Errorf("to see help text")
	}

	switch os.Args[1] {
	case "help", "h", "":
		cmdHelp()
		return nil
	case "new", "n":
		if err := cmdNew(); err != nil {
			return err
		}
		return nil
	case "list", "l":
		if err := cmdList(); err != nil {
			return err
		}
		return nil
	case "add", "a":
		if err := cmdAdd(); err != nil {
			return err
		}
		return nil
	case "done", "d":
		if err := cmdDone(); err != nil {
			return err
		}
		return nil
	case "open", "o":
		if err := cmdOpen(); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("'%s' is not a td command", os.Args[1])
}

func cmdHelp() {
	txt := `Name:
   td - super simple TODO management tool

Usage:
   td command [options] [arguments]

Commands:
     help, h    show help
     new,  n    create config file in current dir
     list, l    list todo
     add,  a    add a new task
     done, d    make a task status done
     open, o    make a task status open`

	fmt.Println(txt)
}

func cmdNew() error {
	if fileExists(filename) {
		return fmt.Errorf("%s is already exists", filename)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(`[]`)); err != nil {
		return fmt.Errorf("can't write to %s: %w", filename, err)
	}

	fmt.Printf("%s was generated\n", filename)
	return nil
}

func cmdList() error {
	if !fileExists(filename) {
		return fmt.Errorf("%s is not found", filename)
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	td, err := newTodo(f)
	if err != nil {
		return err
	}

	if len(td) == 0 {
		return errors.New("no tasks are registered")
	}

	var isAll bool
	if len(os.Args) > 2 && os.Args[2] == "-a" {
		isAll = true
	}

	var txt string
	for _, t := range td {
		if !isAll && t.Status == taskStatusDone {
			continue
		}
		txt += fmt.Sprintf("%s\n", t.line())
	}
	fmt.Print(txt)
	return nil
}

func cmdAdd() error {
	if !fileExists(filename) {
		return fmt.Errorf("%s is not found", filename)
	}

	f, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}
	defer f.Close()

	td, err := newTodo(f)
	if err != nil {
		return err
	}

	t := scanTask(os.Stdin)
	t.ID = td.LastID() + 1
	td = append(td, t)

	b, err := json.Marshal(td)
	if err != nil {
		return fmt.Errorf("can't unmarshal json: %w", err)
	}

	if _, err := f.WriteAt(b, 0); err != nil {
		return fmt.Errorf("can't write to %s: %w", filename, err)
	}
	return nil
}

func cmdDone() error {
	if err := changeTaskStatus(taskStatusDone); err != nil {
		return err
	}
	return nil
}

func cmdOpen() error {
	if err := changeTaskStatus(taskStatusOpen); err != nil {
		return err
	}
	return nil
}

func changeTaskStatus(ts taskStatus) error {
	if !fileExists(filename) {
		return fmt.Errorf("%s is not found", filename)
	}

	f, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}
	defer f.Close()

	td, err := newTodo(f)
	if err != nil {
		return err
	}

	if len(os.Args) <= 2 {
		return errors.New("please enter task id")
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return errors.New("please enter task id in integer")
	}
	td, err = td.changeStatusByID(id, ts)
	if err != nil {
		return err
	}

	b, err := json.Marshal(td)
	if err != nil {
		return fmt.Errorf("can't unmarshal json: %w", err)
	}

	if _, err := f.WriteAt(b, 0); err != nil {
		return fmt.Errorf("can't write to %s: %w", filename, err)
	}
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

type task struct {
	ID       int          `json:"id"`
	Current  bool         `json:"current"`
	Status   taskStatus   `json:"status"`
	Priority taskPriority `json:"priority"`
	Body     string       `json:"body"`
	Children todo         `json:"children"`
}

func scanTask(r io.Reader) task {
	t := task{Status: taskStatusOpen}

	sc := bufio.NewScanner(r)

	for {
		fmt.Print("task's detail:  ")
		if sc.Scan() {
			txt := sc.Text()
			if txt == "" {
				fmt.Println("please write something")
				continue
			}
			t.Body = txt
			break
		}
	}

	for {
		fmt.Print("priority [H/L]: ")
		if sc.Scan() {
			p, err := getTaskPriorityByStr(sc.Text())
			if err != nil {
				fmt.Println("please enter correct priority (H or M or L)")
				continue
			}
			t.Priority = p
			break
		}
	}

	return t
}

func (t task) line() string {
	return fmt.Sprintf("%s %d [%s] %s", t.Priority, t.ID, t.Status, t.Body)
}

type taskStatus int

const (
	_ taskStatus = iota
	taskStatusOpen
	taskStatusDone
)

func (ts taskStatus) String() string {
	switch ts {
	case taskStatusOpen:
		return " "
	case taskStatusDone:
		return "x"
	}
	return "?"
}

type taskPriority int

const (
	_ taskPriority = iota
	taskPriorityLow
	taskPriorityHigh
)

func (tp taskPriority) String() string {
	switch tp {
	case taskPriorityHigh:
		return "H"
	case taskPriorityLow:
		return " "
	}
	return "?"
}

func getTaskPriorityByStr(priorityStr string) (taskPriority, error) {
	switch priorityStr {
	case "H":
		return taskPriorityHigh, nil
	case "L", "":
		return taskPriorityLow, nil
	}
	return 0, fmt.Errorf("task priority not found: %s", priorityStr)
}

type todo []task

func newTodo(r io.Reader) (todo, error) {
	var td todo
	if err := json.NewDecoder(r).Decode(&td); err != nil {
		return nil, fmt.Errorf("can't parse to todo file")
	}
	return td, nil
}

func (td todo) LastID() int {
	var maxID int
	for _, t := range td {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID
}

func (td todo) changeStatusByID(id int, status taskStatus) (todo, error) {
	for i := range td {
		if td[i].ID == id {
			td[i].Status = status
			return td, nil
		}
	}
	return nil, fmt.Errorf("task is not found, id: %d", id)
}
