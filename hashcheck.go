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
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hashcheck <file> [...]")
		return
	}
	var sumOk, sumBad, sumSkip, sumErr int = 0, 0, 0, 0
	for _, file := range os.Args[1:] {
		exhash, err := extractHash(file)
		if err != nil {
			fmt.Printf("SKIP:  %s\n", file)
			sumSkip++
			continue
		}
		hash, err := hashFile(file)
		if err != nil {
			fmt.Printf("ERR:   %s (%v)\n", file, err)
			sumErr++
			continue
		}
		if hash == exhash {
			fmt.Printf("OK:    %s\n", file)
			sumOk++
		} else {
			fmt.Printf("BAD:   %s\n", file)
			sumBad++
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
