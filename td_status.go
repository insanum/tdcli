
package td

import (
	"time"
	"sync"
)

var statusWG sync.WaitGroup
var statusChan chan string

func statusWorker() {
	lastMsgPosted := time.Time{}

	defer statusWG.Done()

	for {

	select {
	case msg, ok := <-statusChan:

		if !ok {
			return /* all done */
		}

		if !lastMsgPosted.IsZero() {
			/* wait small amount of time for previous msg */
			for time.Now().Sub(lastMsgPosted).Seconds() < 2 {
				time.Sleep(1 * time.Second)
			}
		}

		_, h := TermSize()
		TermPrintWidth((h - 1), ColorFmt(msg, StatusFg, StatusBg, StatusAttr))
		TermDrawScreen()

		lastMsgPosted = time.Now()

	case <-time.After(1 * time.Second):

		if !lastMsgPosted.IsZero() &&
		   time.Now().Sub(lastMsgPosted).Seconds() > 10 {
			_, h := TermSize()
			TermClearLines((h - 1), 1)
			lastMsgPosted = time.Time{}
		}
	}

	} /* for */
}

func StatusLines() int {
	return 1
}

func StatusMessage(msg string) {
	statusChan <- msg
}

func StatusStart() {
	statusWG.Add(1)
	statusChan = make(chan string, 20)
	go statusWorker()
}

func StatusStop() {
	close(statusChan)
	statusWG.Wait()
}

