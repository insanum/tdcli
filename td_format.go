
package main

import (
	"strings"
	"strconv"
	"regexp"
	"log"
	"fmt"
	"time"
)

type ClrSpec struct {
	fg   int
	bg   int
	attr int
}
type FmtClr struct {
	idx  int
	clr  ClrSpec
}
type FmtStr struct {
	str    string
	clrs   []FmtClr
	tabbed bool
}

func AttrToTerm(attr string) int {
	switch strings.ToUpper(attr) {
	default:
		fallthrough
	case "0":
		return AttrNone
	case "B":
		return AttrBold
	case "U":
		return AttrUnderline
	case "R":
		return AttrReverse
	}
}

func EmptyFmt(str string) FmtStr {
	return FmtStr{str, []FmtClr{}, false}
}

func ColorFmt(str  string,
	      fg   int,
	      bg   int,
	      attr int) FmtStr {
	return FmtStr{str,
		      []FmtClr{{0, ClrSpec{fg, bg, attr}}},
		      false}
}

func DefaultFmt(str string) FmtStr {
	return FmtStr{str,
		      []FmtClr{{0, ClrSpec{ColorDefault, ColorDefault, AttrNone}}},
		      false}
}

func DefaultFmtAttrReverse(str string) FmtStr {
	return FmtStr{str,
		      []FmtClr{{0, ClrSpec{ColorDefault, ColorDefault, AttrReverse}}},
		      false}
}

func TaskHeaderFormat(title string) FmtStr {
	rxpFmt, err := regexp.Compile(`{([a-z][a-z]?)}`)
	if err != nil {
		log.Fatal(err)
	}

	str := TaskFmtHeader
	res := rxpFmt.FindAllStringSubmatchIndex(str, -1)
	values := make([]interface{}, len(res))

	for i := (len(res) - 1); i >= 0; i-- {
		begin, end := str[:res[i][0]], str[res[i][1]:]
		switch str[res[i][0]:res[i][1]] {
		case "{t}":
			str = begin + "s" + end
			values[i] = title
		}
	}

	str = fmt.Sprintf(str, values...)

	rxpClr, err := regexp.Compile(`{#([0-9]+),([0-9]+),([BUR0])}|{(#-)}`)
	if err != nil {
		log.Fatal(err)
	}

	res = rxpClr.FindAllStringSubmatchIndex(str, -1)
	clrs := []FmtClr{}

	var fg   int = ColorDefault
	var bg   int = ColorDefault
	var attr int = AttrNone

	for {
		res = rxpClr.FindAllStringSubmatchIndex(str, 1)
		if len(res) == 0 {
			break
		}

		begin, end := str[:res[0][0]], str[res[0][1]:]
		seq := str[res[0][0]:res[0][1]]

		if seq == "{#-}" {
			fg   = ColorDefault
			bg   = ColorDefault
			attr = AttrNone
		} else if seq[:2] == "{#" {
			fg, _ = strconv.Atoi(str[res[0][2]:res[0][3]])
			bg, _ = strconv.Atoi(str[res[0][4]:res[0][5]])
			attr = AttrToTerm(str[res[0][6]:res[0][7]])
		}

		clrs = append(clrs, FmtClr{res[0][0], ClrSpec{fg, bg, attr}})
		str = begin + end
	}

	return FmtStr{str, clrs, false}
}

func (task *TDTask) Format() FmtStr {
	rxpFmt, err := regexp.Compile(`{(\*)}|{([a-z][a-z]?)}`)
	if err != nil {
		log.Fatal(err)
	}

	str := TaskFmt
	res := rxpFmt.FindAllStringSubmatchIndex(str, -1)
	values := make([]interface{}, len(res))

	for i := (len(res) - 1); i >= 0; i-- {
		begin, end := str[:res[i][0]], str[res[i][1]:]
		seq := str[res[i][0]:res[i][1]]

		switch seq {
		case "{i}":
			str = begin + "d" + end
			values[i] = task.ID
		case "{t}":
			str = begin + "s" + end
			values[i] = task.Title
		case "{*}":
			str = begin + "s" + end
			values[i] = ""
			if task.Star != 0 {
				values[i] = "*"
			}
		case "{p}":
			str = begin + "s" + end
			switch task.Priority {
			default:
				fallthrough
			case 0:
				values[i] = "0 Low"
			case 1:
				values[i] = "1 Medium"
			case 2:
				values[i] = "2 High"
			case 3:
				values[i] = "3 Top"
			case -1:
				values[i] = "-1 Negative"
			}
		case "{f}":
			str = begin + "s" + end
			values[i] = TDFolders.IdToName(task.Folder)
		case "{l}":
			str = begin + "s" + end
			values[i] = TDLocations.IdToName(task.Location)
		case "{c}":
			str = begin + "s" + end
			values[i] = TDContexts.IdToName(task.Context)
		case "{g}":
			str = begin + "s" + end
			values[i] = TDGoals.IdToName(task.Goal)
		case "{pi}":
			str = begin + "d" + end
			values[i] = task.Parent
		case "{pt}":
			/* XXX Implement a task.IdtoIdx() func */
			str = begin + "" + end
		case "{nc}":
			str = begin + "d" + end
			values[i] = task.Children
		case "{dd}":
			str = begin + "s" + end
			values[i] = ""
			if task.DueDate != 0 {
				values[i] = time.Unix(task.DueDate, 0).Format("2006/1/2")
			}
		case "{dt}":
			str = begin + "s" + end
			values[i] = ""
			if task.DueTime != 0 {
				values[i] = time.Unix(task.DueTime, 0).Format("15:04")
			}
		case "{dm}":
			str = begin + "s" + end
			switch task.DueDateMod {
			default:
				fallthrough
			case 0:
				values[i] = ""
			case 1:
				values[i] = "="
			case 2:
				values[i] = ">"
			case 3:
				values[i] = "?"
			}
		case "{sd}":
			str = begin + "s" + end
			values[i] = ""
			if task.StartDate != 0 {
				values[i] = time.Unix(task.StartDate, 0).Format("2006/1/2")
			}
		case "{st}":
			str = begin + "s" + end
			values[i] = ""
			if task.StartTime != 0 {
				values[i] = time.Unix(task.StartTime, 0).Format("15:04")
			}
		case "{r}":
			str = begin + "d" + end
			values[i] = task.Remind
		case "{rp}":
			str = begin + "s" + end
			values[i] = task.Repeat
		case "{s}":
			str = begin + "s" + end
			switch task.Status {
			default:
				fallthrough
			case 0:
				values[i] = "None"
			case 1:
				values[i] = "Next Action"
			case 2:
				values[i] = "Active"
			case 3:
				values[i] = "Planning"
			case 4:
				values[i] = "Delegated"
			case 5:
				values[i] = "Waiting"
			case 6:
				values[i] = "Hold"
			case 7:
				values[i] = "Postponed"
			case 8:
				values[i] = "Someday"
			case 9:
				values[i] = "Canceled"
			case 10:
				values[i] = "Reference"
			}
		case "{lg}":
			str = begin + "d" + end
			values[i] = task.Length
		case "{md}":
			str = begin + "s" + end
			values[i] = ""
			if task.Modified != 0 {
				values[i] = time.Unix(task.Modified, 0).Format("2006/1/2 15:04")
			}
		case "{cd}":
			str = begin + "s" + end
			values[i] = ""
			if task.Completed != 0 {
				values[i] = time.Unix(task.Completed, 0).Format("2006/1/2 15:04")
			}
		case "{n}":
			str = begin + "s" + end
			values[i] = task.Note
		}
	}

	str = fmt.Sprintf(str, values...)

	rxpClr, err := regexp.Compile(`{#([0-9]+),([0-9]+),([BUR0])}|{(#-)}`)
	if err != nil {
		log.Fatal(err)
	}

	res = rxpClr.FindAllStringSubmatchIndex(str, -1)
	clrs := []FmtClr{}

	var fg   int = ColorDefault
	var bg   int = ColorDefault
	var attr int = AttrNone

	for {
		res = rxpClr.FindAllStringSubmatchIndex(str, 1)
		if len(res) == 0 {
			break
		}

		begin, end := str[:res[0][0]], str[res[0][1]:]
		seq := str[res[0][0]:res[0][1]]

		if seq == "{#-}" {
			fg   = ColorDefault
			bg   = ColorDefault
			attr = AttrNone
		} else if seq[:2] == "{#" {
			fg, _ = strconv.Atoi(str[res[0][2]:res[0][3]])
			bg, _ = strconv.Atoi(str[res[0][4]:res[0][5]])
			attr = AttrToTerm(str[res[0][6]:res[0][7]])
		}

		clrs = append(clrs, FmtClr{res[0][0], ClrSpec{fg, bg, attr}})
		str = begin + end
	}

	return FmtStr{str, clrs, false}
}

func (note *TDNote) Format() FmtStr {
	rxpFmt, err := regexp.Compile(`{([a-z][a-z]?)}`)
	if err != nil {
		log.Fatal(err)
	}

	str := NoteFmt
	res := rxpFmt.FindAllStringSubmatchIndex(str, -1)
	values := make([]interface{}, len(res))

	for i := (len(res) - 1); i >= 0; i-- {
		begin, end := str[:res[i][0]], str[res[i][1]:]
		switch str[res[i][0]:res[i][1]] {
		case "{i}":
			str = begin + "d" + end
			values[i] = note.ID
		case "{t}":
			str = begin + "s" + end
			values[i] = note.Title
		case "{f}":
			str = begin + "s" + end
			values[i] = TDFolders.IdToName(note.Folder)
		case "{md}":
			str = begin + "s" + end
			values[i] = ""
			if note.Modified != 0 {
				values[i] = time.Unix(note.Modified, 0).Format("2006/1/2 15:04")
			}
		case "{ad}":
			str = begin + "s" + end
			values[i] = ""
			if note.Added != 0 {
				values[i] = time.Unix(note.Added, 0).Format("2006/1/2 15:04")
			}
		case "{n}":
			str = begin + "s" + end
			values[i] = note.Text
		}
	}

	str = fmt.Sprintf(str, values...)

	rxpClr, err := regexp.Compile(`{#([0-9]+),([0-9]+),([BUR0])}|{(#-)}`)
	if err != nil {
		log.Fatal(err)
	}

	res = rxpClr.FindAllStringSubmatchIndex(str, -1)
	clrs := []FmtClr{}

	var fg   int = ColorDefault
	var bg   int = ColorDefault
	var attr int = AttrNone

	for {
		res = rxpClr.FindAllStringSubmatchIndex(str, 1)
		if len(res) == 0 {
			break
		}

		begin, end := str[:res[0][0]], str[res[0][1]:]
		seq := str[res[0][0]:res[0][1]]

		if seq == "{#-}" {
			fg   = ColorDefault
			bg   = ColorDefault
			attr = AttrNone
		} else if seq[:2] == "{#" {
			fg, _ = strconv.Atoi(str[res[0][2]:res[0][3]])
			bg, _ = strconv.Atoi(str[res[0][4]:res[0][5]])
			attr = AttrToTerm(str[res[0][6]:res[0][7]])
		}

		clrs = append(clrs, FmtClr{res[0][0], ClrSpec{fg, bg, attr}})
		str = begin + end
	}

	return FmtStr{str, clrs, false}
}

func HelpHeaderFormat(title string) FmtStr {
	rxpFmt, err := regexp.Compile(`{([a-z][a-z]?)}`)
	if err != nil {
		log.Fatal(err)
	}

	str := HelpFmtHeader
	res := rxpFmt.FindAllStringSubmatchIndex(str, -1)
	values := make([]interface{}, len(res))

	for i := (len(res) - 1); i >= 0; i-- {
		begin, end := str[:res[i][0]], str[res[i][1]:]
		switch str[res[i][0]:res[i][1]] {
		case "{t}":
			str = begin + "s" + end
			values[i] = title
		}
	}

	str = fmt.Sprintf(str, values...)

	rxpClr, err := regexp.Compile(`{#([0-9]+),([0-9]+),([BUR0])}|{(#-)}`)
	if err != nil {
		log.Fatal(err)
	}

	res = rxpClr.FindAllStringSubmatchIndex(str, -1)
	clrs := []FmtClr{}

	var fg   int = ColorDefault
	var bg   int = ColorDefault
	var attr int = AttrNone

	for {
		res = rxpClr.FindAllStringSubmatchIndex(str, 1)
		if len(res) == 0 {
			break
		}

		begin, end := str[:res[0][0]], str[res[0][1]:]
		seq := str[res[0][0]:res[0][1]]

		if seq == "{#-}" {
			fg   = ColorDefault
			bg   = ColorDefault
			attr = AttrNone
		} else if seq[:2] == "{#" {
			fg, _ = strconv.Atoi(str[res[0][2]:res[0][3]])
			bg, _ = strconv.Atoi(str[res[0][4]:res[0][5]])
			attr = AttrToTerm(str[res[0][6]:res[0][7]])
		}

		clrs = append(clrs, FmtClr{res[0][0], ClrSpec{fg, bg, attr}})
		str = begin + end
	}

	return FmtStr{str, clrs, false}
}

func HelpKeybindFormat(name  string,
		       key   string,
		       descr string) FmtStr {
	rxpFmt, err := regexp.Compile(`{([a-z][a-z]?)}`)
	if err != nil {
		log.Fatal(err)
	}

	str := HelpFmtKeybind
	res := rxpFmt.FindAllStringSubmatchIndex(str, -1)
	values := make([]interface{}, len(res))

	for i := (len(res) - 1); i >= 0; i-- {
		begin, end := str[:res[i][0]], str[res[i][1]:]
		switch str[res[i][0]:res[i][1]] {
		case "{n}":
			str = begin + "s" + end
			values[i] = name
		case "{k}":
			str = begin + "s" + end
			values[i] = key
		case "{d}":
			str = begin + "s" + end
			values[i] = descr
		}
	}

	str = fmt.Sprintf(str, values...)

	rxpClr, err := regexp.Compile(`{#([0-9]+),([0-9]+),([BUR0])}|{(#-)}`)
	if err != nil {
		log.Fatal(err)
	}

	res = rxpClr.FindAllStringSubmatchIndex(str, -1)
	clrs := []FmtClr{}

	var fg   int = ColorDefault
	var bg   int = ColorDefault
	var attr int = AttrNone

	for {
		res = rxpClr.FindAllStringSubmatchIndex(str, 1)
		if len(res) == 0 {
			break
		}

		begin, end := str[:res[0][0]], str[res[0][1]:]
		seq := str[res[0][0]:res[0][1]]

		if seq == "{#-}" {
			fg   = ColorDefault
			bg   = ColorDefault
			attr = AttrNone
		} else if seq[:2] == "{#" {
			fg, _ = strconv.Atoi(str[res[0][2]:res[0][3]])
			bg, _ = strconv.Atoi(str[res[0][4]:res[0][5]])
			attr = AttrToTerm(str[res[0][6]:res[0][7]])
		}

		clrs = append(clrs, FmtClr{res[0][0], ClrSpec{fg, bg, attr}})
		str = begin + end
	}

	return FmtStr{str, clrs, false}
}

