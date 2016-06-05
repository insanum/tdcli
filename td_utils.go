
package td

import (
	"log"
	"bufio"
	"io/ioutil"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"encoding/gob"
	"encoding/json"
)

func GetUserInput(prompt string) string {
	var input string

	_, h := TermSize()
	TermPrintWidth((h - 1), ColorFmt(prompt, 0, 63, AttrNone))
	TermSetCursor(len(prompt) + 2, (h - 1))
	TermDrawScreen()

	cells := TermSaveCellBuffer()
	TermClose()

	_, err := fmt.Scanln(&input)
	input = strings.TrimSpace(input)
	if err != nil {
		input = ""
	}

	TermInit()
	TermRestoreCellBuffer(cells)
	TermSync()

	return input
}

func ExecCommand(command string, args ...string) {
	cells := TermSaveCellBuffer()
	TermClearScreen(true)

	TermClose()

	cmd := exec.Command(command, args...)
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	TermInit()

	TermRestoreCellBuffer(cells)
	TermSync()
}

func TempFileWrite(data string) (fname string) {
	tmpf, err := ioutil.TempFile("", "td")
	if err != nil {
		log.Fatal(err)
	}

	defer tmpf.Close()

	bio := bufio.NewWriter(tmpf)

	_, err = bio.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}

	bio.Flush()

	fname = tmpf.Name()
	return
}

func TempFileRead(fname string) (data string) {
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	data = string(buf)
	return
}

func TempFileDelete(fname string) {
	os.Remove(fname)
}

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func FileOpen(name     string,
	      truncate bool) *os.File {
	flags := (os.O_RDWR | os.O_APPEND | os.O_CREATE)
	if truncate {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(name, flags, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func FileClose(f *os.File) {
	f.Close()
}

func FileWrite(name string,
	       data interface{}) {
	f := FileOpen(name, true)

	encoder := gob.NewEncoder(f)

	err := encoder.Encode(data)
	if err != nil {
		log.Fatal(err)
	}

	FileClose(f)
}

func FileRead(name string,
	      data interface{}) {
	f := FileOpen(name, false)

	decoder := gob.NewDecoder(f)

	err := decoder.Decode(data)
	if err != nil {
		log.Fatal(err)
	}

	FileClose(f)
}

func DumpJSON(json_data interface{}) {
	data_indent, err := json.MarshalIndent(json_data, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(data_indent)
	fmt.Print("\n")
}

func ValidName(name string) string {
	if name == "" || name == "none" {
		return ""
	} else {
		return name
	}
}

func IdxString(cur_idx  int,
	       num_elem int) string {
	return fmt.Sprintf("(%d/%d)", cur_idx, num_elem)
}

func AskString(prompt string,
	       prime  string) (string, error) {
	_, h := TermSize()
	TermPrintWidth((h - 1), DefaultFmtAttrReverse(prompt))
	TermDrawScreen()

	answer, err := TermInput((len(prompt) + 1), (h - 1), prime)
	TermClearLines((h - 1), 1)
	if err != nil {
		return "", err
	} else {
		return strings.TrimSpace(answer), nil
	}
}

func AskYesNo(prompt string) bool {
	answer, err := AskString(prompt, "")
	if err != nil {
		return false
	}

	if answer == "" {
		return false
	}

	if answer == "y" || answer == "Y" ||
	   answer == "yes" || answer == "Yes" ||
	   answer == "YES" {
		return true
	} else {
		return false
	}
}

