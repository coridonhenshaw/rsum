package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

type QueueEntryStruct struct {
	Filename string
	fi       os.FileInfo
	Hash     string
	err      error
}

var FileQueue chan QueueEntryStruct
var HashQueue chan QueueEntryStruct
var LargeFileMutex sync.RWMutex

func CheckResume(File *QueueEntryStruct) {

	if !ResumeTable.ResumeAvailable {
		return
	}

	var err error

	ResumeTable.Mutex.Lock()

	err = ResumeTable.GetStmt.QueryRow(File.Filename, File.fi.Size(), File.fi.ModTime().Unix()).Scan(&File.Hash)

	ResumeTable.Mutex.Unlock()

	if err != nil {
		File.Hash = ""
	}
}

func Worker(wg *sync.WaitGroup) {

	for File := range FileQueue {
		var Large bool = File.fi.Size() > LargeFileThreshold

		if Large {
			LargeFileMutex.Lock()
		} else {
			LargeFileMutex.RLock()
		}

		CheckResume(&File)
		if len(File.Hash) == 0 {
			if Verbose {
				fmt.Printf("%s\n", File.Filename)
			}
			HashFile(&File)
		} else {
			if Verbose {
				fmt.Printf("Resumed: %s\n", File.Filename)
			}
		}

		if Large {
			LargeFileMutex.Unlock()
		} else {
			LargeFileMutex.RUnlock()
		}

		if File.err == nil {
			HashQueue <- File
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", File.err)
		}
	}

	wg.Done()
}

var ResumeTable *ResumeTableStruct

func CreateSumList(Input []string, Recurse bool, Output string) {

	var err error

	ResumeTable, err = DB.CreateResumeTable("resume_" + strconv.Itoa(HashType))
	if err != nil {
		log.Panic(err)
	}
	defer ResumeTable.Close()

	HashQueue = make(chan QueueEntryStruct, 128)
	var wghash sync.WaitGroup

	wghash.Add(1)
	go func() {

		fh, err := os.OpenFile(Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Panic(err)
		}
		defer fh.Close()

		for item := range HashQueue {
			if !UseBase64 {
				_, err = fmt.Fprintf(fh, "%x  %s\n", item.Hash, item.Filename)
				if err != nil {
					log.Panic(err)
				}
			} else {
				Enc := base64.RawURLEncoding.EncodeToString([]byte(item.Hash))
				_, err = fmt.Fprintf(fh, "%s  %s\n", Enc, item.Filename)
				if err != nil {
					log.Panic(err)
				}
			}

			ResumeTable.Mutex.Lock()

			ResumeTable.AddStmt.Exec(item.Filename, item.fi.Size(), item.fi.ModTime().Unix(), item.Hash)

			ResumeTable.Mutex.Unlock()
		}
		wghash.Done()
	}()

	FileQueue = make(chan QueueEntryStruct, 128)
	var wg sync.WaitGroup

	for i := 0; i < WorkerThreads; i++ {
		wg.Add(1)
		go Worker(&wg)
	}

	var statwg sync.WaitGroup
	StatQueue := make(chan string, 1024)
	for i := 0; i < 8; i++ {
		statwg.Add(1)
		go func() {
			for v := range StatQueue {
				var fi os.FileInfo

				fi, err = os.Stat(v)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v: %s\n", err, v)
					continue
				}

				if fi.Mode().IsRegular() {
					AddFile(v, fi)
				} else if fi.Mode().IsDir() {
					AddDirectory(v, Recurse)
				} else {
					fmt.Fprintf(os.Stderr, "Unsuitable filesystem object:%s\n", v)
				}
			}
			statwg.Done()
		}()
	}

	for _, v := range Input {
		StatQueue <- v
	}

	close(StatQueue)
	statwg.Wait()

	close(FileQueue)
	wg.Wait()

	close(HashQueue)
	wghash.Wait()

	return

}

func AddFile(Filename string, fi os.FileInfo) {
	FileQueue <- QueueEntryStruct{Filename: Filename, fi: fi}
}

type WalkTreeCallback struct{}

func (o *WalkTreeCallback) PreDirectory(Path string) (Accept bool) { return true }
func (o *WalkTreeCallback) Directory(Path string) (Continue bool)  { return true }
func (o *WalkTreeCallback) File(Filename string, fi os.FileInfo) (Continue bool) {
	AddFile(Filename, fi)
	return true
}
func (o *WalkTreeCallback) DirectoryError(Path string, err error) (Continue bool)    { return true }
func (o *WalkTreeCallback) NonFileObject(Path string, f os.FileInfo) (Continue bool) { return true }

func AddDirectory(Path string, Recurse bool) {

	var Callback WalkTreeCallback

	WalkTree(Path, true, Recurse, &Callback)
}
