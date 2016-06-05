
package main

/*
 * Based on 'v3' of the Toodledo Developer's API
 * https://api.toodledo.com/3/index.php
 */

import (
	"flag"
	"log"
	"runtime"
	//"fmt"
)

type Page interface {
	MoveCursorDown()
	MoveCursorUp()

	ScrollBodyDown(lines int)
	ScrollBodyUp(lines int)

	NumHeaderLines() int
	UpdateHeader()
	DrawPage()

	NextPage() interface{}
	PrevPage() interface{}

	ChildPage() interface{}
	ParentPage() interface{}
}

type CurrentState struct {
	page Page
}

var state CurrentState

func Toodledo() {
	var page interface{}

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	/* get cached account information and refresh the token */
	TDAccountInit()
	TDOauth() /* will not return for full authentication */

	/* get the latest account info from the server for sync */
	TDGetData("account/get.php", nil, &tdx)

	/* verifiy the cached userid against fetched account info */
	if tda.UserID == "" {
		/* XXX First sync, verify cache doesn't exist?! */
		tda.UserID = tdx.UserID
	} else if tda.UserID != tdx.UserID {
		log.Fatal("ERROR: Invalid userid from previous authentication!\n")
	}

	TermInit()
	defer TermClose()

	/* initialize the logger */
	LogStart()

	/* initialize the user-defined lists */
	UDLInit()

	/* initialize the status message go channel */
	StatusStart()

	/* initialize the data cache and sync with toodledo servers */
	TDCacheInit()
	TDSync()

	/* update our cached account information */
	tda.AccountInfo = tdx

	/* save the new account/cache data */ 
	/* XXX only write if there are changes */
	TDFileAccountWrite()
	TDFileCacheWrite()

	/* start with the task list */
	state.page = &TDTasks
	TDTasks.DrawPage()

mainloop:
	for {

	ev := TermPollEvent()

	if TermIsEventError(ev) {
		log.Fatal(TermEventError(ev))
	}

	if TermEventKeyCmp(ev, TermKeyEscape) ||
	   TermEventKeyCmp(ev, 'Q') {
		break mainloop
	} else if TermEventKeyCmp(ev, TermKeyEnter) {
		page := state.page.ChildPage()
		if page != nil {
			state.page = page.(Page)
			state.page.DrawPage()
		}
	} else if TermEventKeyCmp(ev, 'q') {
		page := state.page.ParentPage()
		if page != nil {
			state.page = page.(Page)
			state.page.DrawPage()
		}
	} else if TermEventKeyCmp(ev, 'j') {
		state.page.MoveCursorDown()
	} else if TermEventKeyCmp(ev, 'k') {
		state.page.MoveCursorUp()
	} else if TermEventKeyCmp(ev, TermKeyCtrlE) {
		state.page.ScrollBodyUp(1)
	} else if TermEventKeyCmp(ev, TermKeyCtrlY) {
		state.page.ScrollBodyDown(1)
	} else if TermEventKeyCmp(ev, TermKeyCtrlD) {
		_, h := TermSize()
		h -= (state.page.NumHeaderLines() + StatusLines() + 1)
		state.page.ScrollBodyUp(h/2)
	} else if TermEventKeyCmp(ev, TermKeyCtrlU) {
		_, h := TermSize()
		h -= (state.page.NumHeaderLines() + StatusLines() + 1)
		state.page.ScrollBodyDown(h/2)
	} else if TermEventKeyCmp(ev, TermKeySpace) {
		_, h := TermSize()
		h -= (state.page.NumHeaderLines() + StatusLines() + 1)
		state.page.ScrollBodyUp(h)
	} else if TermEventKeyCmp(ev, 'b') {
		_, h := TermSize()
		h -= (state.page.NumHeaderLines() + StatusLines() + 1)
		state.page.ScrollBodyDown(h)
	} else if TermEventKeyCmp(ev, '?') {
		state.page = &TDHelp
		state.page.DrawPage()
	} else if TermEventKeyCmp(ev, 'n') {
		state.page = &TDNotes
		state.page.DrawPage()
	} else if TermEventKeyCmp(ev, 't') {
		state.page = &TDTasks
		state.page.DrawPage()
	} else if TermEventKeyCmp(ev, 'f') {
		state.page = &TDFolders
		state.page.DrawPage()
	} else if TermEventKeyCmp(ev, 'l') {
		state.page = &TDLocations
		state.page.DrawPage()
	} else if TermEventKeyCmp(ev, 'c') {
		state.page = &TDContexts
		state.page.DrawPage()
	} else if TermEventKeyCmp(ev, 'g') {
		state.page = &TDGoals
		state.page.DrawPage()
	} else {
		switch state.page.(type) {
		case *TDTaskList:
			if TermEventKeyCmp(ev, '1') {
				TDTasks.sortOrder = SortOrder1
				TDTasks.TaskListSort(true)
			} else if TermEventKeyCmp(ev, '2') {
				TDTasks.sortOrder = SortOrder2
				TDTasks.TaskListSort(true)
			} else if TermEventKeyCmp(ev, '3') {
				TDTasks.sortOrder = SortOrder3
				TDTasks.TaskListSort(true)
			} else if TermEventKeyCmp(ev, '4') {
				TDTasks.sortOrder = SortOrder4
				TDTasks.TaskListSort(true)
			} else if TermEventKeyCmp(ev, 'A') {
				state.page.(*TDTaskList).AddTask(false)
			} else if TermEventKeyCmp(ev, 'C') {
				state.page.(*TDTaskList).AddTask(true)
			} else if TermEventKeyCmp(ev, 'D') {
				state.page.(*TDTaskList).DeleteTask()
			} else if TermEventKeyCmp(ev, 'X') {
				state.page.(*TDTaskList).CompleteTask()
			} else if TermEventKeyCmp(ev, '/') {
				state.page.(*TDTaskList).SearchTasks(true)
			}
		case *TDNoteList:
			if TermEventKeyCmp(ev, 'a') {
				TDNotes.NoteListSortAlpha(true)
			} else if TermEventKeyCmp(ev, 'm') {
				TDNotes.NoteListSortModified(true)
			} else if TermEventKeyCmp(ev, 'A') {
				state.page.(*TDNoteList).AddNote()
			} else if TermEventKeyCmp(ev, 'D') {
				state.page.(*TDNoteList).DeleteNote()
			} else if TermEventKeyCmp(ev, '/') {
				state.page.(*TDNoteList).SearchNotes(true)
			}
		case *UDLList:
			if TermEventKeyCmp(ev, 'A') {
				state.page.(*UDLList).UDLAddName()
			} else if TermEventKeyCmp(ev, 'E') {
				state.page.(*UDLList).UDLEditName()
			} else if TermEventKeyCmp(ev, 'D') {
				state.page.(*UDLList).UDLDeleteName()
			}
		case *TDTask:
			if TermEventKeyCmp(ev, 'J') {
				page = state.page.NextPage()
				if page != nil {
					state.page = page.(*TDTask)
					state.page.DrawPage()
				}
			} else if TermEventKeyCmp(ev, 'K') {
				page = state.page.PrevPage()
				if page != nil {
					state.page = page.(*TDTask)
					state.page.DrawPage()
				}
			} else if TermEventKeyCmp(ev, '*') {
				state.page.(*TDTask).ToggleTaskStar()
			} else if TermEventKeyCmp(ev, 'T') {
				state.page.(*TDTask).EditTaskTitle()
			} else if TermEventKeyCmp(ev, 'P') {
				state.page.(*TDTask).EditTaskPriority()
			} else if TermEventKeyCmp(ev, 'F') {
				state.page.(*TDTask).EditTaskFolder()
			} else if TermEventKeyCmp(ev, 'S') {
				state.page.(*TDTask).EditTaskTag()
			} else if TermEventKeyCmp(ev, 'D') {
				state.page.(*TDTask).EditTaskDue()
			} else if TermEventKeyCmp(ev, 'R') {
				state.page.(*TDTask).EditTaskRemind()
			} else if TermEventKeyCmp(ev, 'L') {
				state.page.(*TDTask).EditTaskLocation()
			} else if TermEventKeyCmp(ev, 'C') {
				state.page.(*TDTask).EditTaskContext()
			} else if TermEventKeyCmp(ev, 'G') {
				state.page.(*TDTask).EditTaskGoal()
			} else if TermEventKeyCmp(ev, 'N') {
				state.page.(*TDTask).EditTaskNote()
			}
		case *TDNote:
			if TermEventKeyCmp(ev, 'J') {
				page = state.page.NextPage()
				if page != nil {
					state.page = page.(*TDNote)
					state.page.DrawPage()
				}
			} else if TermEventKeyCmp(ev, 'K') {
				page = state.page.PrevPage()
				if page != nil {
					state.page = page.(*TDNote)
					state.page.DrawPage()
				}
			} else if TermEventKeyCmp(ev, 'F') {
				state.page.(*TDNote).EditNoteFolder()
			} else if TermEventKeyCmp(ev, 'E') {
				state.page.(*TDNote).EditNoteText()
			} else if TermEventKeyCmp(ev, 'P') {
				state.page.(*TDNote).PagerNote()
			}
		/*
		case PageTypeHelp:
		*/
		default:
		}
	}

	} /* for */

	StatusStop()
}

