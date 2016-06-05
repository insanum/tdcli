
package main

import (
	"fmt"
	"time"
	"strings"
	"encoding/json"
	"log"
	"os"
	"net/url"
	"strconv"
)

var TaskTitle string = "TOODLEDO: TASK"

type TDTask struct {
	Num         int64          `json:"num"`       /* element from '/get.php. */
	Total       int64          `json:"total"`     /* element from '/get.php' */
	ErrorCode   int64          `json:"errorCode"` /* element from '/edit.php' */
	ErrorDesc   string         `json:"errorDesc"` /* element from '/edit.php' */
	Ref         string         `json:"ref"`       /* element from '/edit.php' */

	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Tag         string         `json:"tag"`
	Folder      int64          `json:"folder"`
	Context     int64          `json:"context"`
	Goal        int64          `json:"goal"`
	Location    int64          `json:"location"`
	Parent      int64          `json:"parent"`
	Children    int64          `json:"children"`
	Order       int64          `json:"order"`
	DueDate     int64          `json:"duedate"`
	DueDateMod  int64          `json:"duedatemod"`
	StartDate   int64          `json:"startdate"`
	DueTime     int64          `json:"duetime"`
	StartTime   int64          `json:"starttime"`
	Remind      int64          `json:"remind"`
	Repeat      string         `json:"repeat"`
	Status      int64          `json:"status"`
	Length      int64          `json:"length"`
	Priority    int64          `json:"priority"`
	Star        int64          `json:"star"`
	Modified    int64          `json:"modified"`
	Completed   int64          `json:"completed"`
	Added       int64          `json:"added"`
	Timer       int64          `json:"timer"`
	TimerOn     int64          `json:"timeron"`
	Note        string         `json:"note"`
	Meta        string         `json:"meta"`
	Previous    int64          `json:"previous"`
	Shared      int64          `json:"shared"`
	SharedOwner int64          `json:"sharedowner"`
	SharedWith  []int64        `json:"sharedowner"`
	AddedBy     int64          `json:"addedby"`
	Via         int64          `json:"via"`
	Attachment  []TDAttachment `json:"attachment"`

	idx         int
	sb          ScrollBuffer
	sbLines     []ScrollLine
}

type TDTaskDeleted struct {
	Num   int64 `json:"num"` /* first element 'tasks/deleted.php. */
	ID    int64 `json:"id"`
	Stamp int64 `json:"stamp"`
}

func TDTaskSyncTask(idx  int,
		    data map[string]string) {
	var tasks []TDTask

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	v := url.Values{}
	v.Set("tasks", string(b))

	TDGetData("tasks/edit.php", v, &tasks)

	if  tasks[0].ErrorCode != 0 {
		StatusMessage(tasks[0].ErrorDesc)
		return
	}

	if tasks[0].ID == TDTasks.tasks[idx].ID {
		TDTasks.tasks[idx].Modified = tasks[0].Modified
		StatusMessage("Task updated")
		TDFileCacheWrite()
	} else {
		StatusMessage("ERROR: Unknown response")
	}
}

func (task *TDTask) TaskSyncStar() {
	data := map[string]string{
		    "id":   strconv.FormatInt(task.ID, 10),
		    "star": strconv.FormatInt(task.Star, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncTitle() {
	data := map[string]string{
		    "id":    strconv.FormatInt(task.ID, 10),
		    "title": task.Title }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncPriority() {
	data := map[string]string{
		    "id":       strconv.FormatInt(task.ID, 10),
		    "priority": strconv.FormatInt(task.Priority, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncFolder() {
	data := map[string]string{
		    "id":     strconv.FormatInt(task.ID, 10),
		    "folder": strconv.FormatInt(task.Folder, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncTag() {
	data := map[string]string{
		    "id":  strconv.FormatInt(task.ID, 10),
		    "tag": task.Tag}
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncDue() {
	data := map[string]string{
		    "id":         strconv.FormatInt(task.ID, 10),
		    "duedate":    strconv.FormatInt(task.DueDate, 10),
		    "duetime":    strconv.FormatInt(task.DueTime, 10),
		    "duedatemod": strconv.FormatInt(task.DueDateMod, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncRemind() {
	data := map[string]string{
		    "id":     strconv.FormatInt(task.ID, 10),
		    "remind": strconv.FormatInt(task.Remind, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncLocation() {
	data := map[string]string{
		    "id":       strconv.FormatInt(task.ID, 10),
		    "location": strconv.FormatInt(task.Location, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncContext() {
	data := map[string]string{
		    "id":      strconv.FormatInt(task.ID, 10),
		    "context": strconv.FormatInt(task.Context, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncGoal() {
	data := map[string]string{
		    "id":   strconv.FormatInt(task.ID, 10),
		    "goal": strconv.FormatInt(task.Goal, 10) }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) TaskSyncNote() {
	data := map[string]string{
		    "id":   strconv.FormatInt(task.ID, 10),
		    "note": task.Note }
	TDTaskSyncTask(task.idx, data)
}

func (task *TDTask) NumHeaderLines() int {
	return 1
}

func (task *TDTask) NumBodyLines() int {
	return len(task.sbLines)
}

func (task *TDTask) MoveCursorDown() {
	task.sb.CursorDown()
}

func (task *TDTask) MoveCursorUp() {
	task.sb.CursorUp()
}

func (task *TDTask) ScrollBodyDown(lines int) {
	task.sb.ScrollDown(lines)
}

func (task *TDTask) ScrollBodyUp(lines int) {
	task.sb.ScrollUp(lines)
}

func (task *TDTask) UpdateHeader() {
	TermDrawScreen()
}

func (task *TDTask) DrawPage() {
	due := time.Unix(task.GetDue(), 0).UTC().Format("2006/1/2 15:04")

	t := fmt.Sprintf(/* 01 */ "%-10s %s\n" +
			 /* 02 */ "%-10s %d\n" +
			 /* 03 */ "%-10s %s\n" +
			 /* 04 */ "%-10s %s\n" +
			 /* 05 */ "%-10s %s\n" +
			 /* 06 */ "%-10s %d\n" +
			 /* 07 */ "%-10s %s\n" +
			 /* 08 */ "%-10s %s\n" +
			 /* 09 */ "%-10s %s\n" +
			 /* 10 */ "%-10s\n"    +
			 /* 11 */ "%s",
			 "Title:",    task.Title,
			 "Priority:", task.Priority,
			 "Folder:",   ValidName(TDFolders.IdToName(task.Folder)),
			 "Tag:",      task.Tag,
			 "Due:",      due,
			 "Remind:",   task.Remind,
			 "Context:",  ValidName(TDContexts.IdToName(task.Context)),
			 "Location:", ValidName(TDLocations.IdToName(task.Location)),
			 "Goal:",     ValidName(TDGoals.IdToName(task.Goal)),
			 "Note:",     task.Note)

	lines := strings.Split(t, "\n")
	task.sbLines = make([]ScrollLine, len(lines))
	for i := 0; i < len(lines); i++ {
		task.sbLines[i].text = EmptyFmt(lines[i])
		task.sbLines[i].selectable = true
		task.sbLines[i].realIdx = i
	}

	task.sb.Init(task.NumHeaderLines(),
		     StatusLines(),
		     &task.sbLines)

	TermClearScreen(false)

	TermTitle(0, TaskTitle, "")

	task.sb.Draw()
}

func (task *TDTask) NextPage() interface{} {
	lastIdx := TDTasks.sbLastSelectIdx
	if lastIdx < (len(TDTasks.sbLines) - 1) {
		TDTasks.sbLastSelectIdx = (lastIdx + 1)
		return &TDTasks.tasks[TDTasks.sbLines[lastIdx + 1].realIdx]
	}
	return nil
}

func (task *TDTask) PrevPage() interface{} {
	lastIdx := TDTasks.sbLastSelectIdx
	if lastIdx > 0 {
		TDTasks.sbLastSelectIdx = (lastIdx - 1)
		return &TDTasks.tasks[TDTasks.sbLines[lastIdx - 1].realIdx]
	}
	return nil
}

func (task *TDTask) ChildPage() interface{} {
	return nil
}

func (task *TDTask) ParentPage() interface{} {
	return &TDTasks
}

func (task *TDTask) GetDue() int64 {
	if task.DueTime != 0 {
		return task.DueTime
	} else if task.DueDate != 0 {
		return task.DueDate
	} else {
		return 0
	}
}

func (task *TDTask) ToggleTaskStar() {

	if task.Star == 0 {
		task.Star = 1
	} else {
		task.Star = 0
	}

	task.TaskSyncStar()
	task.DrawPage()
}

func (task *TDTask) EditTaskTitle() {
	title, err := AskString("Title:", "")
	if err != nil {
		return
	}

	if title == "" {
		return
	}

	task.Title = title

	task.TaskSyncTitle()
	task.DrawPage()
}

func (task *TDTask) EditTaskPriority() {
	prio, err := AskString("Priority [(T)op,(H)igh,(M)edium,(L)ow,(N)egative]:", "")
	if err != nil {
		return
	}

	if prio == "" {
		if !AskYesNo("Clear priority [y/n]?:") {
			return
		}

		prio = "L"
	}

	nprio := 0
	switch prio {
	case "T", "t", "top", "Top", "TOP":
		nprio = 3
	case "H", "h", "high", "High", "HIGH":
		nprio = 2
	case "M", "m", "medium", "Medium", "MEDIUM":
		nprio = 1
	case "L", "l", "low", "Low", "LOW":
		nprio = 0
	case "N", "n", "negative", "Negative", "NEGATIVE":
		nprio = -1
	default:
		StatusMessage("ERROR: Invalid priority")
	}

	task.Priority = int64(nprio)

	task.TaskSyncPriority()
	task.DrawPage()
}

func (task *TDTask) EditTaskFolder() {
	folder, err := AskString("Folder:", "")
	if err != nil {
		return
	}

	if folder == "" {
		if !AskYesNo("Clear folder [y/n]?:") {
			return
		}

		folder = "none"
	}

	if folder == "none" {
		task.Folder = 0
	} else {
		id := TDFolders.NameToId(folder)
		if id == 0 {
			StatusMessage("ERROR: Invalid folder name")
			return
		}

		task.Folder = id
	}

	task.TaskSyncFolder()
	task.DrawPage()
}

func (task *TDTask) EditTaskTag() {
	tag, err := AskString("Tag (comma separated):", task.Tag)
	if err != nil {
		return
	}

	if tag == "" {
		if !AskYesNo("Clear tag [y/n]?:") {
			return
		}
	}

	/* XXX allow editing existing tag instead of overwriting */

	task.Tag = tag

	task.TaskSyncTag()
	task.DrawPage()

}

func (task *TDTask) EditTaskDue() {
	var ddm int64
	var t time.Time

	due, err := AskString("Due date/time:", "")
	if err != nil {
		return
	}

	if due == "" {
		if !AskYesNo("Clear due date/time [y/n]?:") {
			return
		}

		task.DueDate    = 0
		task.DueTime    = 0
		task.DueDateMod = 0
		goto setdue
	}

	ddm = 0 /* due by */
	switch due[0] {
	case '=': /* due on */
		ddm = 1
		due = due[1:]
	case '>': /* due after */
		ddm = 2
		due = due[1:]
	case '?': /* optionally */
		ddm = 3
		due = due[1:]
	default:
	}

	/* golang time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", ...) */

	t, err = time.Parse("2006/1/2", due) /* UTC */
	if err == nil {
		task.DueDate    = t.Unix()
		task.DueTime    = 0
		task.DueDateMod = ddm
		goto setdue

	}

	t, err = time.Parse("2006/1/2 15:04", due) /* UTC */
	if err == nil {
		task.DueDate    = t.Unix()
		task.DueTime    = t.Unix()
		task.DueDateMod = ddm
		goto setdue
	}

	t, err = time.Parse("2006/1/2 3:04pm", due) /* UTC */
	if err == nil {
		task.DueDate    = t.Unix()
		task.DueTime    = t.Unix()
		task.DueDateMod = ddm
		goto setdue
	}

	StatusMessage("ERROR: Invalid due date/time")
	return

setdue:
	task.TaskSyncDue()
	task.DrawPage()
}

func (task *TDTask) EditTaskRemind() {
	remind, err := AskString("Remind (mins):", "")
	if err != nil {
		return
	}

	if remind == "" {
		if !AskYesNo("Clear remind [y/n]?:") {
			return
		}

		remind = "0"
	}

	r, err := strconv.ParseInt(remind, 10, 64)
	if err != nil {
		StatusMessage("ERROR: Invalid remind time")
		return
	}

	if tda.AccountInfo.Pro == 0 {
		switch r {
		case 0, 60:
			/* ok */
		default:
			StatusMessage("ERROR: Invalid remind time")
			return
		}
	} else {
		switch r {
		case 0, 1, 15, 30, 45, 60, 90, 120,
		     180, 240, 1440, 2880, 4320, 5760,
		     7200, 8640, 10080, 20160, 43200:
			/* ok */
		default:
			StatusMessage("ERROR: Invalid remind time")
			return
		}
	}

	task.Remind = r

	task.TaskSyncRemind()
	task.DrawPage()
}

func (task *TDTask) EditTaskLocation() {
	location, err := AskString("Location:", "")
	if err != nil {
		return
	}

	if location == "" {
		if !AskYesNo("Clear location [y/n]?:") {
			return
		}

		location = "none"
	}

	if location == "none" {
		task.Location = 0
	} else {
		id := TDLocations.NameToId(location)
		if id == 0 {
			StatusMessage("ERROR: Invalid location name")
			return
		}

		task.Location = id
	}

	task.TaskSyncLocation()
	task.DrawPage()
}

func (task *TDTask) EditTaskContext() {
	context, err := AskString("Context:", "")
	if err != nil {
		return
	}

	if context == "" {
		if !AskYesNo("Clear context [y/n]?:") {
			return
		}

		context = "none"
	}

	if context == "none" {
		task.Context = 0
	} else {
		id := TDContexts.NameToId(context)
		if id == 0 {
			StatusMessage("ERROR: Invalid context name")
			return
		}

		task.Context = id
	}

	task.TaskSyncContext()
	task.DrawPage()
}

func (task *TDTask) EditTaskGoal() {
	goal, err := AskString("Goal:", "")
	if err != nil {
		return
	}

	if goal == "" {
		if !AskYesNo("Clear goal [y/n]?:") {
			return
		}

		goal = "none"
	}

	if goal == "none" {
		task.Goal = 0
	} else {
		id := TDGoals.NameToId(goal)
		if id == 0 {
			StatusMessage("ERROR: Invalid goal name")
			return
		}

		task.Goal = id
	}

	task.TaskSyncGoal()
	task.DrawPage()
}

func (task *TDTask) EditTaskNote() {
	editor := os.ExpandEnv("$EDITOR")
	if editor == "" {
		StatusMessage("$EDITOR not defined")
		return
	}

	fname := TempFileWrite(task.Note)

	ExecCommand(editor, fname)

	newNote := TempFileRead(fname)
	TempFileDelete(fname)

	if strings.TrimSpace(newNote) ==
	   strings.TrimSpace(task.Note) {
		/* nothing changed */
		StatusMessage("Cancelled")
		return
	}

	task.Note = newNote

	task.TaskSyncNote()
	task.DrawPage()
}

