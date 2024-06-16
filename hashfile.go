package main

import (
	"crypto"
	"hash"
	"io"
	"log"
	"os"
	"sync"

	_ "crypto/subtle"

	_ "golang.org/x/crypto/blake2b"
	// _ "golang.org/x/crypto/sha3"
)

var HashType int = 0

const HashTypeBlake2B = 0
const HashTypeSHA512 = 1
const HashTypeSHA3_512 = 2

func HashConstructor() (h hash.Hash) {

	var err error

	switch HashType {
	case HashTypeBlake2B:
		//		h, err = blake2b.New(32, nil)
		h = crypto.BLAKE2b_512.New()
	case HashTypeSHA512:
		h = crypto.SHA512.New()
	case HashTypeSHA3_512:
		h = crypto.SHA3_512.New()
	}

	if err != nil {
		log.Panic(err)
	}

	return h
}

func HashFile(File *QueueEntryStruct) {

	h := HashConstructor()

	fh, err := os.Open(File.Filename)
	if err != nil {
		File.err = err
		return
	}
	defer fh.Close()

	var sz = int(File.fi.Size())
	//
	// Single read (fast) path for files <= than LargeFileThreshold
	//
	if sz <= LargeFileThreshold {
		var BytesRead int
		var Buffer = make([]byte, sz)

		BytesRead, err = fh.Read(Buffer)
		if BytesRead != sz {
			Buffer = Buffer[:BytesRead]
		}

		h.Write(Buffer)

		if err == io.EOF || err == nil {
			File.Hash = string(h.Sum(nil))
		} else if err != nil {
			File.err = err
			File.Hash = ""
		}
		return
	}

	//
	// Chunked multithreaded path for files > LargeFileThreshold
	//

	var BlockQueue chan (*[]byte)
	BlockQueue = make(chan *[]byte, 4)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for block := range BlockQueue {
			h.Write(*block)
		}
		wg.Done()
		return
	}()

	for {

		var BytesRead int
		var Buffer = make([]byte, LargeFileThreshold)

		BytesRead, err = fh.Read(Buffer)
		if BytesRead != LargeFileThreshold {
			Buffer = Buffer[:BytesRead]
		}

		if err == io.EOF {
			break
		} else if err != nil {
			File.err = err
			break
		}

		BlockQueue <- &Buffer
	}

	close(BlockQueue)

	wg.Wait()

	File.Hash = string(h.Sum(nil))

	return
}
