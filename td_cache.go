
package main

import (
	"sync"
	"time"
)

type ToodledoCache struct {
	Folders     []UDL
	Locations   []UDL
	Contexts    []UDL
	Goals       []UDL
	Notes       []TDNote
	Tasks       []TDTask
	initialized bool
}

var tdc ToodledoCache

func TDSync() {
	var wg sync.WaitGroup
	wg.Add(6)

	full_sync := false
	if !tdc.initialized {
		full_sync = true
	}

	go func() {
		defer wg.Done()
		TDFolders.UDLSync(full_sync)
		TDFolders.UDLListSortAlpha()
	}()

	go func() {
		defer wg.Done()
		TDLocations.UDLSync(full_sync)
		TDLocations.UDLListSortAlpha()
	}()

	go func() {
		defer wg.Done()
		TDContexts.UDLSync(full_sync)
		TDContexts.UDLListSortAlpha()
	}()

	go func() {
		defer wg.Done()
		TDGoals.UDLSync(full_sync)
		TDGoals.UDLListSortAlpha()
	}()

	go func() {
		defer wg.Done()
		TDNotesSync(full_sync)
		TDNotes.NoteListSortModified(false)
	}()

	go func() {
		defer wg.Done()
		TDTasksSync(full_sync)
		TDTasks.sortOrder = SortOrder1
		TDTasks.TaskListSort(false)
	}()

	//TDListsSync(full_sync)
	//TDOutlinesSync(full_sync)
	//TDHabitsSync(full_sync)

	wg.Wait()
	tda.LastSync = time.Now()
}

func TDFileCacheExists() bool {
	return FileExists(CacheFileName)
}

func TDFileCacheWrite() {
	tdc.Folders = make([]UDL, len(TDFolders.udls))
	copy(tdc.Folders, TDFolders.udls)

	tdc.Locations = make([]UDL, len(TDLocations.udls))
	copy(tdc.Locations, TDLocations.udls)

	tdc.Contexts = make([]UDL, len(TDContexts.udls))
	copy(tdc.Contexts, TDContexts.udls)

	tdc.Goals = make([]UDL, len(TDGoals.udls))
	copy(tdc.Goals, TDGoals.udls)

	tdc.Notes = make([]TDNote, len(TDNotes.notes))
	copy(tdc.Notes, TDNotes.notes)

	tdc.Tasks = make([]TDTask, len(TDTasks.tasks))
	copy(tdc.Tasks, TDTasks.tasks)

	FileWrite(CacheFileName, tdc)
}

func TDFileCacheRead() {
	FileRead(CacheFileName, &tdc)

	TDFolders.udls = make([]UDL, len(tdc.Folders))
	copy(TDFolders.udls, tdc.Folders)

	TDLocations.udls = make([]UDL, len(tdc.Locations))
	copy(TDLocations.udls, tdc.Locations)

	TDContexts.udls = make([]UDL, len(tdc.Contexts))
	copy(TDContexts.udls, tdc.Contexts)

	TDGoals.udls = make([]UDL, len(tdc.Goals))
	copy(TDGoals.udls, tdc.Goals)

	TDNotes.notes = make([]TDNote, len(tdc.Notes))
	copy(TDNotes.notes, tdc.Notes)

	TDTasks.tasks = make([]TDTask, len(tdc.Tasks))
	copy(TDTasks.tasks, tdc.Tasks)
}

func TDCacheInit() {
	tdc.initialized = false
	if TDFileCacheExists() {
		TDFileCacheRead()
		tdc.initialized = true
	}
	//DumpJSON(tdc)
}

