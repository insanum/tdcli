
package td

import (
	"net/url"
	//"fmt"
	"strconv"
	"sort"
	"time"
	"os"
	"strings"
	"regexp"
	"log"
	"encoding/json"
)

var TaskListTitle string = "TOODLEDO: TASKS"

var TaskTemplate string =
`Title: <title>
Priority:
Folder:
Tag:
Due:
Remind:
Location:
Context:
Goal:
---
<task note>
`

type TDTaskList struct {
	tasks           []TDTask
	sbLastSelectIdx int
	sortOrder       [3]string
	sb              ScrollBuffer
	sbLines         []ScrollLine
}

var TDTasks TDTaskList

func TDTasksAddTask(title    string,
		    priority string,
		    folder   string,
		    tag      string,
		    due      string,
		    remind   string,
		    location string,
		    context  string,
		    goal     string,
		    note     string,
		    parent   int64) {
	var tasks []TDTask

	/*
	 * REQUIRED:
	 * title
	 *
	 * OPTIONAL:
	 * folder
	 * context
	 * goal
	 * location
	 * priority
	 * status
	 * star
	 * duration
	 * remind
	 * starttime
	 * duetime
	 * completed
	 * duedatemod
	 * repeat
	 * tag
	 * duedate
	 * startdate
	 * note
	 * parent
	 * meta
	 */
	data := map[string]string{
		    "title":    title,
		    "parent":   strconv.FormatInt(parent, 10),
		    "priority": priority,
		    "folder":   strconv.FormatInt(TDFolders.NameToId(folder), 10),
		    "tag":      tag,
		    //"duedate":  due,
		    //"duetime":  due,
		    //"remind":   remind,
		    "location": strconv.FormatInt(TDLocations.NameToId(location), 10),
		    "context":  strconv.FormatInt(TDContexts.NameToId(location), 10),
		    "goal":     strconv.FormatInt(TDGoals.NameToId(location), 10),
		    "note":     note }

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	v := url.Values{}
	v.Set("tasks", string(b))

	TDGetData("tasks/add.php", v, &tasks)

	if  tasks[0].ErrorCode != 0 {
		StatusMessage(tasks[0].ErrorDesc)
		return
	}

	tasks[0].Parent     = parent
	//tasks[0].Priority   = priority
	tasks[0].Tag        = tag
	//tasks[0].DueDate    =
	//tasks[0].DueTime    =
	//tasks[0].DueDateMod =
	tasks[0].Folder     = TDFolders.NameToId(location)
	tasks[0].Location   = TDLocations.NameToId(location)
	tasks[0].Context    = TDContexts.NameToId(location)
	tasks[0].Goal       = TDGoals.NameToId(location)
	tasks[0].Note       = note

	TDTasks.tasks = append(TDTasks.tasks, tasks[0])
	StatusMessage("Note added")
	TDFileCacheWrite()
	TDTasks.TaskListSort(true)
}

func TDTasksDeleteTask(idx int) {
	var tasks []TDTask

	data := []string{ strconv.FormatInt(TDTasks.tasks[idx].ID, 10) }
	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	v := url.Values{}
	v.Set("tasks", string(b))

	TDGetData("tasks/delete.php", v, &tasks)

	if  tasks[0].ErrorCode != 0 {
		StatusMessage(tasks[0].ErrorDesc)
		return
	}

	if tasks[0].ID == TDTasks.tasks[idx].ID {
		TDTasks.tasks = append(TDTasks.tasks[:idx],
				       TDTasks.tasks[idx+1:]...)
		StatusMessage("Task deleted")
		TDFileCacheWrite()
		TDTasks.TaskListSort(true)
	} else {
		StatusMessage("ERROR: Unknown response")
	}
}

func TDTasksCompleteTask(idx int) {
	var tasks []TDTask

	TDTasks.tasks[idx].Completed = time.Now().UTC().Unix()

	data := map[string]string{
		    "id":         strconv.FormatInt(TDTasks.tasks[idx].ID, 10),
		    "completed":  strconv.FormatInt(TDTasks.tasks[idx].Completed, 10),
		    "reschedule": "1" }

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
		TDTasks.tasks[idx].Modified  = tasks[0].Modified
		TDTasks.tasks[idx].Completed = tasks[0].Completed
		StatusMessage("Task completed")
		if tasks[0].Completed == 0 {
			/* XXX
			 * Fetch the single task info instead of
			 * a full sync here...
			 */
			TDTasksSync(true)
		} else {
			TDTasksSync(false)
		}
		TDFileCacheWrite()
		TDTasks.TaskListSort(true)
	} else {
		StatusMessage("ERROR: Unknown response")
	}
}

/*
 * Sort task list alphabetically by title.
 */
type TaskListSortByAlpha []TDTask
func (tasks TaskListSortByAlpha) Len() int {
	return len(tasks)
}
func (tasks TaskListSortByAlpha) Swap(i int,
				      j int) {
	tasks[i], tasks[j] = tasks[j], tasks[i]
}
func (tasks TaskListSortByAlpha) Less(i int,
				      j int) bool {
	return (strings.ToLower(tasks[i].Title) <
		strings.ToLower(tasks[j].Title))
}
func AlphaEqual(t1 *ScrollLine,
		t2 *ScrollLine) bool {
	return (strings.ToLower(TDTasks.tasks[t1.realIdx].Title) ==
		strings.ToLower(TDTasks.tasks[t2.realIdx].Title))
}
func AlphaCompare(t1 *ScrollLine,
		  t2 *ScrollLine) bool {
	return (strings.ToLower(TDTasks.tasks[t1.realIdx].Title) <
		strings.ToLower(TDTasks.tasks[t2.realIdx].Title))
}
type TaskListSortByAlphaSbLines []ScrollLine
func (sbLines TaskListSortByAlphaSbLines) Len() int {
	return len(sbLines)
}
func (sbLines TaskListSortByAlphaSbLines) Swap(i int,
					       j int) {
	sbLines[i], sbLines[j] = sbLines[j], sbLines[i]
}
func (sbLines TaskListSortByAlphaSbLines) Less(i int,
					       j int) bool {
	return AlphaCompare(&sbLines[i], &sbLines[j])
}
func TaskListSortByAlphaSbLinesFunc(sbLines []ScrollLine) {
	sort.Sort(TaskListSortByAlphaSbLines(sbLines))
}

/*
 * Sort task list by last modified date.
 */
type TaskListSortByModified []TDTask
func (tasks TaskListSortByModified) Len() int {
	return len(tasks)
}
func (tasks TaskListSortByModified) Swap(i int,
					 j int) {
	tasks[i], tasks[j] = tasks[j], tasks[i]
}
func (tasks TaskListSortByModified) Less(i int,
					 j int) bool {
	return tasks[i].Modified > tasks[j].Modified
}
func ModifiedEqual(t1 *ScrollLine,
		   t2 *ScrollLine) bool {
	return (TDTasks.tasks[t1.realIdx].Modified ==
		TDTasks.tasks[t2.realIdx].Modified)
}
func ModifiedCompare(t1 *ScrollLine,
		     t2 *ScrollLine) bool {
	return (TDTasks.tasks[t1.realIdx].Modified >
		TDTasks.tasks[t2.realIdx].Modified)
}
type TaskListSortByModifiedSbLines []ScrollLine
func (sbLines TaskListSortByModifiedSbLines) Len() int {
	return len(sbLines)
}
func (sbLines TaskListSortByModifiedSbLines) Swap(i int,
						  j int) {
	sbLines[i], sbLines[j] = sbLines[j], sbLines[i]
}
func (sbLines TaskListSortByModifiedSbLines) Less(i int,
						  j int) bool {
	return ModifiedCompare(&sbLines[i], &sbLines[j])
}
func TaskListSortByModifiedSbLinesFunc(sbLines []ScrollLine) {
	sort.Sort(TaskListSortByModifiedSbLines(sbLines))
}

/*
 * Sort task list by due date.
 */
type TaskListSortByDue []TDTask
func (tasks TaskListSortByDue) Len() int {
	return len(tasks)
}
func (tasks TaskListSortByDue) Swap(i int,
				    j int) {
	tasks[i], tasks[j] = tasks[j], tasks[i]
}
func (tasks TaskListSortByDue) Less(i int,
				    j int) bool {
	i_time := tasks[i].GetDue()
	j_time := tasks[j].GetDue()

	/* no due date is always last */
	if j_time == 0 {
		return true
	}
	if i_time == 0 {
		return false
	}

	/* reverse sort as we want older due dates at the top */
	return i_time < j_time
}
func DueEqual(t1 *ScrollLine,
	      t2 *ScrollLine) bool {
	t1_time := TDTasks.tasks[t1.realIdx].GetDue()
	t2_time := TDTasks.tasks[t2.realIdx].GetDue()

	return (t1_time == t2_time)
}
func DueCompare(t1 *ScrollLine,
		t2 *ScrollLine) bool {
	t1_time := TDTasks.tasks[t1.realIdx].GetDue()
	t2_time := TDTasks.tasks[t2.realIdx].GetDue()

	/* no due date is always last */
	if t2_time == 0 {
		return true
	}
	if t1_time == 0 {
		return false
	}

	/* reverse sort as we want older due dates at the top */
	return t1_time < t2_time
}
type TaskListSortByDueSbLines []ScrollLine
func (sbLines TaskListSortByDueSbLines) Len() int {
	return len(sbLines)
}
func (sbLines TaskListSortByDueSbLines) Swap(i int,
					     j int) {
	sbLines[i], sbLines[j] = sbLines[j], sbLines[i]
}
func (sbLines TaskListSortByDueSbLines) Less(i int,
					     j int) bool {
	return DueCompare(&sbLines[i], &sbLines[j])
}
func TaskListSortByDueSbLinesFunc(sbLines []ScrollLine) {
	sort.Sort(TaskListSortByDueSbLines(sbLines))
}

/*
 * Sort task list by priority.
 */
type TaskListSortByPriority []TDTask
func (tasks TaskListSortByPriority) Len() int {
	return len(tasks)
}
func (tasks TaskListSortByPriority) Swap(i int,
					 j int) {
	tasks[i], tasks[j] = tasks[j], tasks[i]
}
func (tasks TaskListSortByPriority) Less(i int,
					 j int) bool {
	return tasks[i].Priority > tasks[j].Priority
}
func PriorityEqual(t1 *ScrollLine,
		   t2 *ScrollLine) bool {
	return (TDTasks.tasks[t1.realIdx].Priority ==
		TDTasks.tasks[t2.realIdx].Priority)
}
func PriorityCompare(t1 *ScrollLine,
		     t2 *ScrollLine) bool {
	return (TDTasks.tasks[t1.realIdx].Priority >
		TDTasks.tasks[t2.realIdx].Priority)
}
type TaskListSortByPrioritySbLines []ScrollLine
func (sbLines TaskListSortByPrioritySbLines) Len() int {
	return len(sbLines)
}
func (sbLines TaskListSortByPrioritySbLines) Swap(i int,
						  j int) {
	sbLines[i], sbLines[j] = sbLines[j], sbLines[i]
}
func (sbLines TaskListSortByPrioritySbLines) Less(i int,
						  j int) bool {
	return PriorityCompare(&sbLines[i], &sbLines[j])
}
func TaskListSortByPrioritySbLinesFunc(sbLines []ScrollLine) {
	sort.Sort(TaskListSortByPrioritySbLines(sbLines))
}

/*
 * Sort task list by Folder (folder by alpha).
 */
type TaskListSortByFolder []TDTask
func (tasks TaskListSortByFolder) Len() int {
	return len(tasks)
}
func (tasks TaskListSortByFolder) Swap(i int,
				       j int) {
	tasks[i], tasks[j] = tasks[j], tasks[i]
}
func (tasks TaskListSortByFolder) Less(i int,
				       j int) bool {
	iF := strings.ToLower(TDFolders.IdToName(tasks[i].Folder))
	jF := strings.ToLower(TDFolders.IdToName(tasks[j].Folder))

	/* "none" folder is always last */
	if jF == "none" {
		return true
	}
	if iF == "none" {
		return false
	}

	return iF < jF
}
func FolderEqual(t1 *ScrollLine,
		 t2 *ScrollLine) bool {
	return (strings.ToLower(TDFolders.IdToName(TDTasks.tasks[t1.realIdx].Folder)) ==
		strings.ToLower(TDFolders.IdToName(TDTasks.tasks[t2.realIdx].Folder)))
}
func FolderCompare(t1 *ScrollLine,
		   t2 *ScrollLine) bool {
	iF := strings.ToLower(TDFolders.IdToName(TDTasks.tasks[t1.realIdx].Folder))
	jF := strings.ToLower(TDFolders.IdToName(TDTasks.tasks[t2.realIdx].Folder))

	/* "none" folder is always last */
	if jF == "none" {
		return true
	}
	if iF == "none" {
		return false
	}

	return iF < jF
}
type TaskListSortByFolderSbLines []ScrollLine
func (sbLines TaskListSortByFolderSbLines) Len() int {
	return len(sbLines)
}
func (sbLines TaskListSortByFolderSbLines) Swap(i int,
						j int) {
	sbLines[i], sbLines[j] = sbLines[j], sbLines[i]
}
func (sbLines TaskListSortByFolderSbLines) Less(i int,
						j int) bool {
	return FolderCompare(&sbLines[i], &sbLines[j])
}
func TaskListSortByFolderSbLinesFunc(sbLines []ScrollLine) {
	sort.Sort(TaskListSortByFolderSbLines(sbLines))
}

/*
 * Sort task list by Star.
 */
type TaskListSortByStar []TDTask
func (tasks TaskListSortByStar) Len() int {
	return len(tasks)
}
func (tasks TaskListSortByStar) Swap(i int,
				     j int) {
	tasks[i], tasks[j] = tasks[j], tasks[i]
}
func (tasks TaskListSortByStar) Less(i int,
				     j int) bool {
	return tasks[i].Star > tasks[j].Star
}
func StarEqual(t1 *ScrollLine,
	       t2 *ScrollLine) bool {
	return (TDTasks.tasks[t1.realIdx].Star ==
		TDTasks.tasks[t2.realIdx].Star)
}
func StarCompare(t1 *ScrollLine,
		 t2 *ScrollLine) bool {
	return (TDTasks.tasks[t1.realIdx].Star >
		TDTasks.tasks[t2.realIdx].Star)
}
type TaskListSortByStarSbLines []ScrollLine
func (sbLines TaskListSortByStarSbLines) Len() int {
	return len(sbLines)
}
func (sbLines TaskListSortByStarSbLines) Swap(i int,
					      j int) {
	sbLines[i], sbLines[j] = sbLines[j], sbLines[i]
}
func (sbLines TaskListSortByStarSbLines) Less(i int,
					      j int) bool {
	return StarCompare(&sbLines[i], &sbLines[j])
}

func TaskListSortByStarSbLinesFunc(sbLines []ScrollLine) {
	sort.Sort(TaskListSortByStarSbLines(sbLines))
}

func (tasks *TDTaskList) TaskListSetIndexes() {
	for i := 0; i < len(tasks.tasks); i++ {
		tasks.tasks[i].idx = i
	}
}

func (tasks *TDTaskList) SegsByDue() {
	now := time.Now()
	dueSegs := [...]struct{
		label  string
		cutoff time.Time
	}{
		{ "Due in the Past",
			time.Date(now.Year(), now.Month(), now.Day(),
				  0, 0, 0, 0, now.Location()) },
		{ "Due Today",
			time.Date(now.Year(), now.Month(), now.Day(),
				  0, 0, 0, 0, now.Location()).
			AddDate(0, 0, 1) },
		{ "Due Tomorrow",
			time.Date(now.Year(), now.Month(), now.Day(),
				  0, 0, 0, 0, now.Location()).
			AddDate(0, 0, 2) },
		{ "Due This Week",
			time.Date(now.Year(), now.Month(), now.Day(),
				  0, 0, 0, 0, now.Location()).
			AddDate(0, 0, int(time.Saturday - now.Weekday() + 1)) },
		{ "Due This Month",
			time.Date(now.Year(), now.Month(), 1,
				  0, 0, 0, 0, now.Location()).
			AddDate(0, 1, 0) },
		{ "Due Next Month",
			time.Date(now.Year(), now.Month() + 1, 1,
				  0, 0, 0, 0, now.Location()).
			AddDate(0, 2, 0) },
		{ "Due in the Future",
			time.Date(now.Year(), now.Month() + 2, 1,
				  0, 0, 0, 0, now.Location()).
			AddDate(0, 3, 0) },
		{ "No Due Date", time.Time{} },
	}

	var line ScrollLine
	i := 0

	for k := range dueSegs {
		segHasContent := false

		line.text = TaskHeaderFormat(dueSegs[k].label)
		line.selectable = false

		tasks.sbLines = append(tasks.sbLines, ScrollLine{})
		copy(tasks.sbLines[i+1:], tasks.sbLines[i:])
		tasks.sbLines[i] = line
		i++

		for i < len(tasks.sbLines) {
			td := tasks.tasks[tasks.sbLines[i].realIdx].GetDue()
			if td == 0 {
				// make sure on final dueSeg
				if !dueSegs[k].cutoff.IsZero() {
					break
				}
			} else {
				d := time.Unix(td, 0).UTC()
				if d.After(dueSegs[k].cutoff) {
					break
				}
			}

			segHasContent = true
			i++
		}

		if !segHasContent {
			// trim the previous segment if empty
			i--
			tasks.sbLines = append(tasks.sbLines[:i], tasks.sbLines[i+1:]...)
		}
	}
}

func (tasks *TDTaskList) SegsByPriority() {
	prios := [...]struct{
		value int64
		label string
	}{
		{  3, "3 Top" },
		{  2, "2 High" },
		{  1, "1 Medium" },
		{  0, "0 Low" },
		{ -1, "-1 Negative" },
	}

	var line ScrollLine
	i := 0

	for k := range prios {
		segHasContent := false

		line.text = TaskHeaderFormat(prios[k].label)
		line.selectable = false

		tasks.sbLines = append(tasks.sbLines, ScrollLine{})
		copy(tasks.sbLines[i+1:], tasks.sbLines[i:])
		tasks.sbLines[i] = line
		i++

		for i < len(tasks.sbLines) {
			if tasks.tasks[tasks.sbLines[i].realIdx].Priority !=
			   prios[k].value {
				break
			}

			segHasContent = true
			i++
		}

		if !segHasContent {
			// trim the previous segment if empty
			i--
			tasks.sbLines = append(tasks.sbLines[:i], tasks.sbLines[i+1:]...)
		}
	}
}

func (tasks *TDTaskList) SegsByFolder() {
	numFolders := TDFolders.NameCount()

	var line ScrollLine
	i := 0

	for k := 0; k < numFolders; k++ {
		segHasContent := false
		folder := tasks.tasks[tasks.sbLines[i].realIdx].Folder

		line.text = DefaultFmtAttrReverse(TDFolders.IdToName(folder))
		line.selectable = false

		tasks.sbLines = append(tasks.sbLines, ScrollLine{})
		copy(tasks.sbLines[i+1:], tasks.sbLines[i:])
		tasks.sbLines[i] = line
		i++

		for i < len(tasks.sbLines) {
			if tasks.tasks[tasks.sbLines[i].realIdx].Folder !=
			   folder {
				break
			}

			segHasContent = true
			i++
		}

		if !segHasContent {
			// trim the previous segment if empty
			i--
			tasks.sbLines = append(tasks.sbLines[:i], tasks.sbLines[i+1:]...)
		}
	}
}

func (tasks *TDTaskList) SegsByStar() {
	stars := [...]struct{
		value int64
		label string
	}{
		{ 1, "(*) Starred" },
		{ 0, "(-) Not Starred" },
	}

	var line ScrollLine
	i := 0

	for k := range stars {
		segHasContent := false

		line.text = TaskHeaderFormat(stars[k].label)
		line.selectable = false

		tasks.sbLines = append(tasks.sbLines, ScrollLine{})
		copy(tasks.sbLines[i+1:], tasks.sbLines[i:])
		tasks.sbLines[i] = line
		i++

		for i < len(tasks.sbLines) {
			if tasks.tasks[tasks.sbLines[i].realIdx].Star !=
			   stars[k].value {
				break
			}

			segHasContent = true
			i++
		}

		if !segHasContent {
			// trim the previous segment if empty
			i--
			tasks.sbLines = append(tasks.sbLines[:i], tasks.sbLines[i+1:]...)
		}
	}
}

func (tasks *TDTaskList) GetNextSeg(i int) (found    bool,
					    startIdx int,
					    endIdx   int) {
	found = false
	startIdx = i
	endIdx = 0

	for startIdx < len(tasks.sbLines) {
		if !tasks.sbLines[startIdx].selectable {
			found = true
			break
		}
		startIdx++
	}

	if !found {
		return
	}

	startIdx++
	endIdx = (startIdx + 1)

	for endIdx < len(tasks.sbLines) {
		if !tasks.sbLines[endIdx].selectable {
			break
		}
		endIdx++
	}

	return
}

type SortCmp  func(*ScrollLine, *ScrollLine) bool
type SortFunc func([]ScrollLine)

func GetSortCmp(sortStr string) SortCmp {
	switch sortStr {
	case "alpha":
		return AlphaEqual
	case "modified":
		return ModifiedEqual
	case "due":
		return DueEqual
	case "priority":
		return PriorityEqual
	case "folder":
		return FolderEqual
	case "star":
		return StarEqual
	default:
		return nil
	}
}

func GetSortAlg(sortStr string) SortFunc {
	switch sortStr {
	case "alpha":
		return TaskListSortByAlphaSbLinesFunc
	case "modified":
		return TaskListSortByModifiedSbLinesFunc
	case "due":
		return TaskListSortByDueSbLinesFunc
	case "priority":
		return TaskListSortByPrioritySbLinesFunc
	case "folder":
		return TaskListSortByFolderSbLinesFunc
	case "star":
		return TaskListSortByStarSbLinesFunc
	default:
		return nil
	}
}

func SortInSeg(sbLines  []ScrollLine,
	       cmpLvl1  SortCmp,
	       cmpLvl2  SortCmp,
	       sortLvl2 SortFunc,
	       cmpLvl3  SortCmp,
	       sortLvl3 SortFunc) {
	var found bool
	var startIdx int
	var endIdx int

	if cmpLvl1 == nil || sortLvl2 == nil {
		return
	}

	/* second level sort */

	startIdx = 0
	endIdx = startIdx + 1

	for endIdx < len(sbLines) {
		found = false

		for (endIdx < len(sbLines)) &&
		    cmpLvl1(&sbLines[endIdx-1], &sbLines[endIdx]) {
			found = true
			endIdx++
		}

		if found {
			/* sort between startIdx and endIdx */
			sortLvl2(sbLines[startIdx:endIdx])
		}

		startIdx = endIdx
		endIdx = startIdx + 1
	}

	if cmpLvl2 == nil || sortLvl3 == nil {
		return
	}

	/* third level sort */

	startIdx = 0
	endIdx = startIdx + 1

	for endIdx < len(sbLines) {
		found = false

		for (endIdx < len(sbLines)) &&
		     (cmpLvl1(&sbLines[endIdx-1], &sbLines[endIdx]) &&
		      cmpLvl2(&sbLines[endIdx-1], &sbLines[endIdx])) {
			found = true
			endIdx++
		}

		if found {
			/* sort between startIdx and endIdx */
			sortLvl3(sbLines[startIdx:endIdx])
		}

		startIdx = endIdx
		endIdx = startIdx + 1
	}
}

func (tasks *TDTaskList) TaskListGroupInsertChildren(children []ScrollLine) {
	for i := (len(children) - 1); i >= 0; i-- {
		/*
		 * Find this child's parent in the scroll buffer list
		 * and insert the child immediately after the parent.
		 * Walking the child list in reverse guarantees the
		 * child ordering under the parent remains.
		 */
		inserted := false

		for j := 0; j < len(tasks.sbLines); j++ {
			if tasks.tasks[children[i].realIdx].Parent ==
			   tasks.tasks[tasks.sbLines[j].realIdx].ID {
				tasks.sbLines = append(tasks.sbLines, ScrollLine{})
				copy(tasks.sbLines[j+2:], tasks.sbLines[j+1:])
				tasks.sbLines[j+1] = children[i]
				inserted = true
				break
			}
		}

		if !inserted {
			log.Fatal("ERROR: Child task without a parent!\n")
		}
	}
}

func (tasks *TDTaskList) TaskListSort(draw bool) {
	switch tasks.sortOrder[0] {
	case "alpha":
		sort.Sort(TaskListSortByAlpha(tasks.tasks))
	case "modified":
		sort.Sort(TaskListSortByModified(tasks.tasks))
	case "due":
		sort.Sort(TaskListSortByDue(tasks.tasks))
	case "priority":
		sort.Sort(TaskListSortByPriority(tasks.tasks))
	case "folder":
		sort.Sort(TaskListSortByFolder(tasks.tasks))
	case "star":
		sort.Sort(TaskListSortByStar(tasks.tasks))
	default:
		msg := "ERROR: Unknown sort order\n"
		LogPrintf(msg)
		if draw {
			StatusMessage(msg)
		}
	}

	tasks.TaskListSetIndexes()

	/* create the first pass of the scroll buffer lines */

	tasks.sbLines = []ScrollLine{}
	children := []ScrollLine{}

	var line ScrollLine
	i := 0

	for i < len(tasks.tasks) {
		if tasks.tasks[i].Completed == 0 {
			line.text = tasks.tasks[i].Format()
			line.selectable = true
			line.realIdx = tasks.tasks[i].idx

			if tasks.tasks[i].Parent == 0 {
				tasks.sbLines = append(tasks.sbLines, line)
			} else {
				line.text.tabbed = true
				children = append(children, line)
			}
		}

		i++
	}

	/* chunk the scroll buffer lines into segments based on the sort */

	hasSegs := false

	switch tasks.sortOrder[0] {
	case "due":
		tasks.SegsByDue()
		hasSegs = true
	case "priority":
		tasks.SegsByPriority()
		hasSegs = true
	case "folder":
		tasks.SegsByFolder()
		hasSegs = true
	case "star":
		tasks.SegsByStar()
		hasSegs = true
	}

	if hasSegs {
		i := 0
		for {
			segExists, startIdx, endIdx := tasks.GetNextSeg(i)
			if !segExists {
				break
			}

			SortInSeg(tasks.sbLines[startIdx:endIdx],
				  GetSortCmp(tasks.sortOrder[0]),
				  GetSortCmp(tasks.sortOrder[1]),
				  GetSortAlg(tasks.sortOrder[1]),
				  GetSortCmp(tasks.sortOrder[2]),
				  GetSortAlg(tasks.sortOrder[2]))

			i = endIdx
		}

		/* sort the children (by second level sort) and insert */
		SortInSeg(children,
			GetSortCmp(tasks.sortOrder[0]),
			GetSortCmp(tasks.sortOrder[1]),
			GetSortAlg(tasks.sortOrder[1]),
			GetSortCmp(tasks.sortOrder[2]),
			GetSortAlg(tasks.sortOrder[2]))
	}

	/* insert the children under their parents */
	tasks.TaskListGroupInsertChildren(children)

	/* finally done... render the scroll buffer */

	tasks.sb.Init(tasks.NumHeaderLines(),
		      StatusLines(),
		      &tasks.sbLines)

	if draw {
		tasks.DrawPage()
	}
}

func (tasks *TDTaskList) SearchTasks(draw bool) {
	search, err := AskString("Search:", "")
	if err != nil {
		return
	}

	if search == "" {
		return
	}

	rxp, err := regexp.Compile(`(?i)` + search)
	if err != nil {
		log.Fatal(err)
	}

	/* pre-sort by priority */
	sort.Sort(TaskListSortByPriority(tasks.tasks))

	tasks.sbLines = []ScrollLine{}
	tasks.sbLastSelectIdx = 0

	for i := 0; i < len(tasks.tasks); i++ {
		if tasks.tasks[i].Completed != 0 {
			continue
		}

		if rxp.FindString(tasks.tasks[i].Title)                          != "" ||
		   rxp.FindString(TDFolders.IdToName(tasks.tasks[i].Folder))     != "" ||
		   rxp.FindString(tasks.tasks[i].Tag)                            != "" ||
		   rxp.FindString(TDLocations.IdToName(tasks.tasks[i].Location)) != "" ||
		   rxp.FindString(TDContexts.IdToName(tasks.tasks[i].Context))   != "" ||
		   rxp.FindString(TDGoals.IdToName(tasks.tasks[i].Goal))         != "" ||
		   rxp.FindString(tasks.tasks[i].Note)                           != "" {
			var line ScrollLine
			line.text = tasks.tasks[i].Format()
			line.selectable = true
			line.realIdx = tasks.tasks[i].idx
			tasks.sbLines = append(tasks.sbLines, line)
		}
	}

	tasks.sb.Init(tasks.NumHeaderLines(),
		      StatusLines(),
		      &tasks.sbLines)

	if draw {
		tasks.DrawPage()
	}
}

func (tasks *TDTaskList) NumHeaderLines() int {
	return 1
}

func (tasks *TDTaskList) NumBodyLines() int {
	return len(tasks.sbLines)
}

func (tasks *TDTaskList) MoveCursorDown() {
	tasks.sb.CursorDown()
}

func (tasks *TDTaskList) MoveCursorUp() {
	tasks.sb.CursorUp()
}

func (tasks *TDTaskList) ScrollBodyDown(lines int) {
	tasks.sb.ScrollDown(lines)
}

func (tasks *TDTaskList) ScrollBodyUp(lines int) {
	tasks.sb.ScrollUp(lines)
}

func (tasks *TDTaskList) UpdateHeader() {
	str := IdxString(((tasks.sb.row - tasks.NumHeaderLines()) + 1),
			 tasks.NumBodyLines())
	TermTitle(0, TaskListTitle, str)

	TermDrawScreen()
}

func (tasks *TDTaskList) TaskStats() (open      int,
				      completed int) {
	open, completed = 0, 0
	for i := 0; i < len(tasks.tasks); i++ {
		if tasks.tasks[i].Completed == 0 {
			open++
		} else {
			completed++
		}
	}
	return
}

func (tasks *TDTaskList) DrawPage() {
	TermClearScreen(false)

	str := IdxString(1, tasks.NumBodyLines())
	TermTitle(0, TaskListTitle, str)

	tasks.sb.Draw()
}

func (tasks *TDTaskList) NextPage() interface{} {
	return nil
}

func (tasks *TDTaskList) PrevPage() interface{} {
	return nil
}

func (tasks *TDTaskList) ChildPage() interface{} {
	idx := tasks.sb.CursorToIdx()
	if tasks.sbLines[idx].selectable {
		tasks.sbLastSelectIdx = idx
		return &tasks.tasks[tasks.sbLines[idx].realIdx]
	} else {
		return nil
	}
}

func (tasks *TDTaskList) ParentPage() interface{} {
	return nil
}

func (tasks *TDTaskList) AddTask(child bool) {
	editor := os.ExpandEnv("$EDITOR")
	if editor == "" {
		StatusMessage("$EDITOR not defined")
		return
	}

	nparent := int64(0)
	if child {
		if tda.AccountInfo.Pro == 0 {
			StatusMessage("ERROR: Unable to create subtask")
			return
		}

		if tasks.tasks[tasks.sbLines[tasks.sb.CursorToIdx()].realIdx].Parent != 0 {
			StatusMessage("ERROR: Not a valid parent task")
			return
		} else {
			nparent = tasks.tasks[tasks.sbLines[tasks.sb.CursorToIdx()].realIdx].ID
		}
	}

	fname := TempFileWrite(TaskTemplate)

	ExecCommand(editor, fname)

	task := TempFileRead(fname)
	TempFileDelete(fname)

	if strings.TrimSpace(task) ==
	   strings.TrimSpace(TaskTemplate) {
		/* nothing changed */
		StatusMessage("Cancelled")
		return
	}

	rxpTitle, err := regexp.Compile(`^\s*Title:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpPriority, err := regexp.Compile(`^\s*Priority:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpFolder, err := regexp.Compile(`^\s*Folder:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpTag, err := regexp.Compile(`^\s*Tag:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpDue, err := regexp.Compile(`^\s*Due:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpRemind, err := regexp.Compile(`^\s*Remind:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpLocation, err := regexp.Compile(`^\s*Location:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpContext, err := regexp.Compile(`^\s*Context:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpGoal, err := regexp.Compile(`^\s*Goal:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpSep, err := regexp.Compile(`^\s*---\s*$`)
	if err != nil {
		log.Fatal(err)
	}

	spltask := strings.SplitN(task, "\n", 11)
	var res []string


	res = rxpTitle.FindStringSubmatch(spltask[0])
	if res == nil {
		StatusMessage("ERROR: Invalid task title")
		return
	}
	ntitle := strings.TrimSpace(res[1])


	res = rxpPriority.FindStringSubmatch(spltask[1])
	if res == nil {
		StatusMessage("ERROR: Invalid task priority")
		return
	}
	npriority := strings.TrimSpace(res[1])
	/* XXX verify priority (ignore or re-edit?) */


	res = rxpFolder.FindStringSubmatch(spltask[2])
	if res == nil {
		StatusMessage("ERROR: Invalid task folder")
		return
	}
	nfolder := strings.TrimSpace(res[1])
	/* XXX verify folder (ignore or re-edit?) */


	res = rxpTag.FindStringSubmatch(spltask[3])
	if res == nil {
		StatusMessage("ERROR: Invalid task tag")
		return
	}
	ntag := strings.TrimSpace(res[1])
	/* XXX verify tag (ignore or re-edit?) */


	res = rxpDue.FindStringSubmatch(spltask[4])
	if res == nil {
		StatusMessage("ERROR: Invalid task due date/time")
		return
	}
	ndue := strings.TrimSpace(res[1])
	/* XXX verify due (ignore or re-edit?) */


	res = rxpRemind.FindStringSubmatch(spltask[5])
	if res == nil {
		StatusMessage("ERROR: Invalid task remind date/time")
		return
	}
	nremind := strings.TrimSpace(res[1])
	/* XXX verify remind (ignore or re-edit?) */


	res = rxpLocation.FindStringSubmatch(spltask[6])
	if res == nil {
		StatusMessage("ERROR: Invalid task location")
		return
	}
	nlocation := strings.TrimSpace(res[1])
	/* XXX verify location (ignore or re-edit?) */


	res = rxpContext.FindStringSubmatch(spltask[7])
	if res == nil {
		StatusMessage("ERROR: Invalid task context")
		return
	}
	ncontext := strings.TrimSpace(res[1])
	/* XXX verify context (ignore or re-edit?) */


	res = rxpGoal.FindStringSubmatch(spltask[8])
	if res == nil {
		StatusMessage("ERROR: Invalid task goal")
		return
	}
	ngoal := strings.TrimSpace(res[1])
	/* XXX verify goal (ignore or re-edit?) */


	res = rxpSep.FindStringSubmatch(spltask[9])
	if res == nil {
		StatusMessage("ERROR: Invalid task format")
		return
	}


	nnote := strings.TrimSpace(spltask[10])


	TDTasksAddTask(ntitle, npriority, nfolder, ntag, ndue,
		       nremind, nlocation, ncontext, ngoal, nnote,
		       nparent)
	tasks.DrawPage()

	return

	/*
inputError:

	return
	*/
}

func (tasks *TDTaskList) DeleteTask() {
	if !AskYesNo("Delete task [y/n]?:") {
		return
	}

	TDTasksDeleteTask(tasks.sbLines[tasks.sb.CursorToIdx()].realIdx)
	tasks.DrawPage()
}

func (tasks *TDTaskList) CompleteTask() {
	if !AskYesNo("Complete task [y/n]?:") {
		return
	}

	TDTasksCompleteTask(tasks.sbLines[tasks.sb.CursorToIdx()].realIdx)
	tasks.DrawPage()
}

/*
 * Sync all tasks from Toodledo servers.
 */
func TDTasksSync(full_sync bool) {
	var after int64
	var v url.Values
	var tasks []TDTask
	var dtasks []TDTaskDeleted

	/*
	 * Get any new or edited notes...
	 */

	if full_sync {
		after = 0
	} else {
		if tda.AccountInfo.LastEditTask >= tdx.LastEditTask {
			goto sync_tasks_deleted
		}

		after = tda.AccountInfo.LastEditTask
	}

	v = url.Values{}
	v.Set("after", strconv.FormatInt(after, 10))
	// id,title,modified,completed
	v.Set("fields", "folder,context,goal,location,tag,startdate,duedate,duedatemod,starttime,duetime,remind,repeat,status,star,priority,length,timer,timeron,added,note,parent,children,order,meta,previous,via")

	TDGetData("tasks/get.php", v, &tasks)

	//fmt.Printf("num=%d total=%d\n", tasks[0].Num, tasks[0].Total)

	if full_sync {
		TDTasks.tasks = tasks[1:]
		return
	}

	for _, t1 := range tasks[1:] {
		updated := false
		for i, t2 := range TDTasks.tasks {
			if t1.ID == t2.ID {
				//fmt.Printf("Updating Task -> %s (%d)\n", t1.Title, t1.ID)
				TDTasks.tasks[i] = t1
				updated = true
			}
		}
		if !updated {
			//fmt.Printf("Adding Task -> %s (%d)\n", t1.Title, t1.ID)
			TDTasks.tasks = append(TDTasks.tasks, t1)
		}
	}

	/*
	 * Get any deleted notes...
	 */

sync_tasks_deleted:

	if tda.AccountInfo.LastDeleteTask >= tdx.LastDeleteTask {
		return
	}

	after = tda.AccountInfo.LastDeleteTask

	v = url.Values{}
	v.Set("after", strconv.FormatInt(after, 10))

	TDGetData("tasks/deleted.php", v, &dtasks)

	//fmt.Printf("num=%d\n", dnotes[0].Num)

	for _, dt1 := range dtasks[1:] {
		for i, dt2 := range TDTasks.tasks {
			if dt1.ID == dt2.ID {
				//fmt.Printf("Deleting Task -> %s (%d)\n", dt2.Title, dt2.ID)
				TDTasks.tasks = append(TDTasks.tasks[:i], TDTasks.tasks[i+1:]...)
			}
		}
	}
}

