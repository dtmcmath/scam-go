package scamutil

import (
	"bufio"
	"unicode/utf8"
)

// FillRuneChannelFromScanner reads the content of the scanner and
// pushes all the runes into the channel (ch).  If there is an error
// reading the file, we return that value.  The channel is always
// closed when the method returns.
func FillRuneChannelFromScanner(scanner *bufio.Scanner, ch chan<- rune) error {
	defer close(ch)

	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		cur := scanner.Text()
		pos := 0
		for r, sz := utf8.DecodeRuneInString(cur[pos:]) ; sz > 0 ; {
			ch <- r
			pos += sz
			r, sz = utf8.DecodeRuneInString(cur[pos:])
		}
	}
	return scanner.Err()
}
