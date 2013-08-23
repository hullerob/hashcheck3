// Copyright 2013 Robert Hülle
// No warranty. WTFPL v2

/*

Hashckeck verifies file with CRC32 included in filename.

example:
	touch file_00000000
	hashcheck file_00000000


limitations:

Last substring of 8 hexadecimal chars in filename is taken as CRC32. This
substring does not have to be separated from rest of filename in any way,
except by non-hexadecimal characters.

*/
package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"unicode"
)

type result struct {
	res  int
	file string
	err  error
}

const (
	SKIP = iota
	OK
	BAD
	ERR
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hashcheck <file> [...]")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	procs := runtime.GOMAXPROCS(0)
	files := make(chan string)
	results := make(chan result)
	wg := new(sync.WaitGroup)
	wg.Add(procs)
	for i := 0; i < procs; i++ {
		go hasher(files, results, wg)
	}
	go args(files, os.Args[1:])
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	var sumOk, sumBad, sumSkip, sumErr int
L1:
	for {
		select {
		case r := <-results:
			switch r.res {
			case SKIP:
				sumSkip++
				fmt.Printf("SKIP: %s\n", r.file)
			case OK:
				sumOk++
				fmt.Printf("OK:   %s\n", r.file)
			case BAD:
				sumBad++
				fmt.Printf("BAD:  %s\n", r.file)
			case ERR:
				sumErr++
				fmt.Printf("ERR:  %s: %v\n", r.file, r.err)
			}
		case _ = <-done:
			break L1
		}
	}
	if sumOk > 0 {
		fmt.Printf("[ok: %d]", sumOk)
	}
	if sumBad > 0 {
		fmt.Printf("[bad: %d]", sumBad)
	}
	if sumSkip > 0 {
		fmt.Printf("[skip: %d]", sumSkip)
	}
	if sumErr > 0 {
		fmt.Printf("[err: %d]", sumErr)
	}
	fmt.Println("")
}

func args(files chan<- string, filenames []string) {
	for _, f := range filenames {
		files <- f
	}
	close(files)
}

func hasher(files <-chan string, res chan<- result, wg *sync.WaitGroup) {
	for file := range files {
		expHash, err := extractHash(file)
		if err != nil {
			res <- result{SKIP, file, err}
			continue
		}
		hash, err := hashFile(file)
		if err != nil {
			res <- result{ERR, file, err}
			continue
		}
		if hash == expHash {
			res <- result{OK, file, nil}
		} else {
			res <- result{BAD, file, nil}
		}
	}
	wg.Done()
}

func extractHash(name string) (hash uint32, err error) {
	var valid bool = false
	hexs := strings.FieldsFunc(name, isNotHex)
	for _, hex := range hexs {
		if len(hex) == 8 {
			fmt.Sscanf(hex, "%x", &hash)
			valid = true
		}
	}
	if !valid {
		err = errors.New("no hexadecimal string")
	}
	return
}

func isNotHex(c rune) bool {
	return !unicode.Is(unicode.ASCII_Hex_Digit, c)
}

func hashFile(fileName string) (uint32, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	buff := make([]byte, 8192)
	hash := crc32.NewIEEE()
	var length int
	for length, err = file.Read(buff); length > 0; length, err = file.Read(buff) {
		hash.Write(buff[:length])
	}
	if err != nil && err != io.EOF {
		return hash.Sum32(), err
	}
	return hash.Sum32(), nil
}
