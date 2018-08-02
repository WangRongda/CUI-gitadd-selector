package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	ui "github.com/gizak/termui"
)

type itemers []struct {
	filePath       string
	children       []uint // if item is a file(leaf), the value is nil; if item is a directory, the value is a list which save the subitem index
	parent         []uint // root is nil, record item's directory item index
	selectedStatus int    // 0 (unselected for file and unselected all file for directory), 1(selected), 2(only for directory status, mean some file selected)
	gitStatus      rune   // 'M'(modified), 'D'(deleted), '?'(untracked)
}

func main() {
	items := getItems()
	list := make([]string, len(items))
	for i := range items {
		list[i] = "  [" + string(items[i].gitStatus) + "] " + items[i].filePath
	}
	items.startUI(list)
	// items.gitadd()
}

func getItems() itemers {
	cmd := `git status --porcelain | grep -E "^.\S+.*$"`
	// output example:
	//MM main.go
	//?? test.go
	// D test2.go
	output, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if nil != err {
		log.Fatal(err)
	}
	outputList := strings.Split(string(output), "\n")
	//一般结尾多了个空""
	if len(outputList) > 0 && "" == outputList[len(outputList)-1] {
		outputList = outputList[:len(outputList)-1]
	}

	items := make(itemers, len(outputList))
	for i, outputline := range outputList {
		items[i].gitStatus = rune(outputline[1])
		items[i].filePath = outputline[3:]
		items[i].selectedStatus = 0
	}
	return items
}

// git add filepaths of selected items
func (i itemers) gitadd() error {
	cmd := `git add`
	for _, v := range i {
		if 1 == v.selectedStatus {
			cmd += " " + v.filePath
		}
	}
	fmt.Println(cmd)
	output, err := exec.Command("/bin/bash", "-c", cmd).Output()
	if nil != err {
		log.Fatal(output, err)
	}
	return nil
}

func (i itemers) selectall() {
	for index := range i {
		i[index].selectedStatus = 1
	}
}

func (i itemers) unselectall() {
	for index := range i {
		i[index].selectedStatus = 0
	}
}

func (i itemers) startUI(strs []string) {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	// strs := []string{
	// 	"[0]Something went wrong",
	// 	"[1] editbox.go",
	// 	"[2] interrupt.go",
	// 	"[3] keyboard.go",
	// 	"[4] output.go",
	// 	"[5] random_out.go",
	// 	"[6] dashboard.go",
	// 	"[7] nsf/termbox-go",
	// 	"[8] editbox.go",
	// 	"[9] interrupt.go",
	// 	"[10] keyboard.go",
	// 	"[11] output.go",
	// 	"[12] random_out.go",
	// 	"[13] dashboard.go",
	// 	"[14] nsf/termbox-go",
	// 	"[15] editbox.go",
	// 	"[16] interrupt.go",
	// 	"[17] keyboard.go",
	// 	"[18] output.go",
	// 	"[19] random_out.go",
	// 	"[20] dashboard.go",
	// }
	l := ui.NewList()
	l.Items = strs
	l.ItemFgColor = ui.ColorBlack
	// l.BorderLabel = "List"
	l.Y = 0
	l.Height = len(strs)
	l.Width = 30
	l.Border = false
	l.ItemBgColor = ui.ColorYellow

	n := 0 //当前行, selected num
	shift := 0
	selectedWrap := "[%s](bg-red,fg-bold)"
	l.Items[0] = fmt.Sprintf(selectedWrap, l.Items[0])
	ui.Render(l) // feel free to call Render, it's async and non-block

	// event handler...
	ui.Handle("/sys/kbd/<space>", func(ui.Event) {
		// press q to quit
		if 0 == i[n].selectedStatus {
			l.Items[n] = "[* " + l.Items[n][3:]
			i[n].selectedStatus = 1
		} else {
			l.Items[n] = "[  " + l.Items[n][3:]
			i[n].selectedStatus = 0
		}
		ui.Render(l)
	})
	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		ui.Close()
		i.gitadd()
		// fmt.Println("finish loop")
		os.Exit(0)
	})
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		// press q to quit
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/j", func(ui.Event) {
		if n >= shift+ui.TermHeight()-2 && shift+ui.TermHeight() < len(strs) {
			shift++
			l.Y = -shift
		} else if n >= len(l.Items)-1 {
			return
		}
		l.Items[n] = l.Items[n][1 : len(l.Items[n])-17]
		n++
		l.Items[n] = fmt.Sprintf(selectedWrap, l.Items[n])
		ui.Render(l)
	})
	ui.Handle("/sys/kbd/k", func(ui.Event) {
		if n == shift+1 && shift > 0 {
			shift--
			l.Y = -shift
		} else if n <= 0 {
			return
		}
		l.Items[n] = l.Items[n][1 : len(l.Items[n])-17]
		n--
		l.Items[n] = fmt.Sprintf(selectedWrap, l.Items[n])
		ui.Render(l)
	})
	ui.Handle("/sys/wnd/resize", func(ui.Event) {
		// ui.Body.Height = ui.TermHeight()
		// ui.Body.Align()
		ui.Clear()
		ui.Render(l)
	})
	ui.Loop() // block until StopLoop is called
}
