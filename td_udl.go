
package td

import (
	"sort"
	"strconv"
	"net/url"
	"fmt"
	"strings"
)

var UDLListTitle string = "TOODLEDO: "

type UDL struct {
	ErrorCode int64  `json:"errorCode"` /* element from '/add.php' */
	ErrorDesc string `json:"errorDesc"` /* element from '/add.php' */
	//Deleted   int64  `json:"deleted"`   /* element from '/delete.php */

	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Archived  int64  `json:"archived"` /* folder status */

	idx       int
}

const (
       FolderUDL = iota
       LocationUDL
       ContextUDL
       GoalUDL
)

type UDLList struct {
	udls    []UDL
	udlType int
	sb      ScrollBuffer
	sbLines []ScrollLine
}

var TDFolders   UDLList
var TDLocations UDLList
var TDContexts  UDLList
var TDGoals     UDLList

func (udls *UDLList) UDLAdd(name string) {
	var udl []UDL

	v := url.Values{}
	v.Set("name", name)

	TDGetData(udls.TypeStr() + "/add.php", v, &udl)

	if  udl[0].ErrorCode != 0 {
		StatusMessage(udl[0].ErrorDesc)
		return
	}

	udls.udls = append(udls.udls, udl[0])
	StatusMessage(fmt.Sprintf("Added '%s'", name))
	TDFileCacheWrite()
	udls.UDLListSortAlpha()
}

func (udls *UDLList) UDLEdit(idx  int,
			     name string) {
	var udl []UDL

	v := url.Values{}
	v.Set("id", strconv.FormatInt(udls.udls[idx].ID, 10))
	v.Set("name", name)

	TDGetData(udls.TypeStr() + "/edit.php", v, &udl)

	if  udl[0].ErrorCode != 0 {
		StatusMessage(udl[0].ErrorDesc)
		return
	}

	if udl[0].ID == udls.udls[idx].ID {
		udls.udls[idx]     = udl[0]
		udls.udls[idx].idx = idx

		StatusMessage(fmt.Sprintf("Updated '%s'", name))
		TDFileCacheWrite()
		udls.UDLListSortAlpha()
	} else {
		StatusMessage("ERROR: Unknown response")
	}
}

func (udls *UDLList) UDLDelete(idx int) {
	var udl []UDL

	v := url.Values{}
	v.Set("id", strconv.FormatInt(udls.udls[idx].ID, 10))

	TDGetData(udls.TypeStr() + "/delete.php", v, &udl)

	/* XXX not getting 'deleted' back...
	if  udl[0].ErrorCode != 0 {
		StatusMessage(udl[0].ErrorDesc)
		return
	}

	if udl[0].Deleted == udls.udls[idx].ID {
		StatusMessage(fmt.Sprintf("Deleted '%s'", udls.udls[idx].Name))
		udls.udls = append(udls.udls[:idx], udls.udls[idx+1:]...)
		TDTasksSync(false)
		TDNotesSync(false)
		TDFileCacheWrite()
		udls.UDLListSortAlpha()
	} else {
		StatusMessage("ERROR: Unknown response")
	}
	*/
	StatusMessage(fmt.Sprintf("Deleted '%s'", udls.udls[idx].Name))
	udls.UDLSync(true)
	TDTasksSync(false)
	TDNotesSync(false)
	TDFileCacheWrite()
	udls.UDLListSortAlpha()
}

/*
 * Sort folder list alphabetically by name.
 */
type UDLListSortByAlpha []UDL
func (udls UDLListSortByAlpha) Len() int {
	return len(udls)
}
func (udls UDLListSortByAlpha) Swap(i int,
				    j int) {
	udls[i], udls[j] = udls[j], udls[i]
}
func (udls UDLListSortByAlpha) Less(i int,
				    j int) bool {
	return udls[i].Name < udls[j].Name
}


func (udls *UDLList) UDLListSetIndexes() {
	for i := 0; i < len(udls.udls); i++ {
		udls.udls[i].idx = i
	}
}

func (udls *UDLList) UDLListSortAlpha() {
	sort.Sort(UDLListSortByAlpha(udls.udls))
	udls.UDLListSetIndexes()
}

func (udls *UDLList) NumHeaderLines() int {
	return 1
}

func (udls *UDLList) NumBodyLines() int {
	return len(udls.sbLines)
}

func (udls *UDLList) MoveCursorDown() {
	udls.sb.CursorDown()
}

func (udls *UDLList) MoveCursorUp() {
	udls.sb.CursorUp()
}

func (udls *UDLList) ScrollBodyDown(lines int) {
	udls.sb.ScrollDown(lines)
}

func (udls *UDLList) ScrollBodyUp(lines int) {
	udls.sb.ScrollUp(lines)
}

func (udls *UDLList) UpdateHeader() {
	strL := UDLListTitle + strings.ToUpper(udls.TypeStr())
	strR := IdxString(((udls.sb.row - udls.NumHeaderLines()) + 1),
			  udls.NumBodyLines())
	TermTitle(0, strL, strR)

	TermDrawScreen()
}

func (udls *UDLList) DrawPage() {
	udls.sbLines = make([]ScrollLine, udls.Count())
	i, j := 0, 0
	for j < len(udls.udls) {
		if udls.udls[j].Archived == 0 {
			udls.sbLines[i].text = EmptyFmt(udls.udls[j].Name)
			udls.sbLines[i].selectable = true
			udls.sbLines[i].realIdx = udls.udls[j].idx
			i++
		}

		j++
	}

	udls.sb.Init(udls.NumHeaderLines(),
		     StatusLines(),
		     &udls.sbLines)

	TermClearScreen(false)

	strL := UDLListTitle + strings.ToUpper(udls.TypeStr())
	strR := IdxString(1, udls.NumBodyLines())
	TermTitle(0, strL, strR)

	udls.sb.Draw()
}

func (udls *UDLList) NextPage() interface{} {
	return nil
}

func (udls *UDLList) PrevPage() interface{} {
	return nil
}

func (udls *UDLList) ChildPage() interface{} {
	return nil
}

func (udls *UDLList) ParentPage() interface{} {
	return nil
}

func (udls *UDLList) TypeStr() string {
	switch udls.udlType {
	case FolderUDL:
		return "folders"
	case LocationUDL:
		return "locations"
	case ContextUDL:
		return "contexts"
	case GoalUDL:
		return "goals"
	default:
		return "unknown"
	}
}

func (udls *UDLList) Count() int {
	count := 0
	for _, f := range udls.udls {
		if f.Archived == 0 {
			count++
		}
	}

	return count
}

func (udls *UDLList) UDLAddName() {
	name, err := AskString("Add name:", "")
	if err != nil {
		return
	}

	if name == "" {
		return
	}

	if udls.NameToId(name) != 0 {
		StatusMessage("ERROR: Already exists")
		return
	}

	udls.UDLAdd(name)
	udls.DrawPage()
}

func (udls *UDLList) UDLEditName() {
	name, err := AskString("New name:", "")
	if err != nil {
		return
	}

	if name == "" {
		return
	}

	if udls.NameToId(name) != 0 {
		StatusMessage("ERROR: Already exists")
		return
	}

	udls.UDLEdit(udls.sb.CursorToIdx(), name)
	udls.DrawPage()
}

func (udls *UDLList) UDLDeleteName() {
	if !AskYesNo("Delete [y/n]?:") {
		return
	}

	udls.UDLDelete(udls.sb.CursorToIdx())
	udls.DrawPage()
}

func (udls *UDLList) UDLSync(full_sync bool) {
	if !full_sync {
		return
	}

	TDGetData(udls.TypeStr() + "/get.php", nil, &udls.udls)
}

func (udls *UDLList) IdToName(id int64) string {
	for _, udl := range udls.udls {
		if id == udl.ID {
			return udl.Name
		}
	}

	return "none"
}

func (udls *UDLList) NameToId(name string) int64 {
	for _, udl := range udls.udls {
		if name == udl.Name {
			return udl.ID
		}
	}

	return 0
}

func (udls *UDLList) NameCount() int {
	return (udls.Count() + 1) /* account for 'none' */
}

func UDLInit() {
	TDFolders.udlType   = FolderUDL
	TDLocations.udlType = LocationUDL
	TDContexts.udlType  = ContextUDL
	TDGoals.udlType     = GoalUDL
}

