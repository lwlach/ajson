package ajson

import (
	"io"
)

type buffer struct {
	data   []byte
	length int
	index  int
}

const (
	quotes    byte = '"'
	backslash byte = '\\'
	skipS     byte = ' '
	skipN     byte = '\n'
	skipR     byte = '\r'
	skipT     byte = '\t'
	bracketL  byte = '['
	bracketR  byte = ']'
	bracesL   byte = '{'
	bracesR   byte = '}'
)

var (
	_null  = []byte("null")
	_true  = []byte("true")
	_false = []byte("false")
)

func newBuffer(body []byte, clone bool) (b *buffer) {
	b = &buffer{
		length: len(body),
	}
	if clone {
		copy(body, b.data)
	} else {
		b.data = body
	}
	return
}

func (b *buffer) first() (c byte, err error) {
	for ; b.index < b.length; b.index++ {
		c = b.data[b.index]
		if !(c == skipS || c == skipR || c == skipN || c == skipT) {
			return c, nil
		}
	}
	return 0, io.EOF
}

func (b *buffer) next() (c byte, err error) {
	if err := b.step(); err != nil {
		return 0, err
	}
	return b.first()
}

func (b *buffer) scan(s byte, skip bool) (from, to int, err error) {
	var c byte
	find := false
	from = b.index
	to = b.index
	for ; b.index < b.length; b.index++ {
		c = b.data[b.index]
		if c == s && !b.backslash() {
			err = b.step()
			return
		}
		if skip && (c == skipS || c == skipR || c == skipN || c == skipT) {
			if !find {
				from++
				to++
			}
		} else {
			find = true
			to++
		}
	}
	return -1, -1, io.EOF
}

func (b *buffer) backslash() (result bool) {
	for i := b.index - 1; i >= 0; i-- {
		if b.data[i] == backslash {
			result = !result
		} else {
			break
		}
	}
	return
}

func (b *buffer) skip(s byte) bool {
	for ; b.index < b.length; b.index++ {
		if b.data[b.index] == s && !b.backslash() {
			return true
		}
	}
	return false
}

func (b *buffer) numeric() error {
	var c byte
	find := 0
	for ; b.index < b.length; b.index++ {
		c = b.data[b.index]
		switch true {
		case c >= '0' && c <= '9':
			find |= 4
		case c == '.':
			if find&2 == 0 {
				find &= 2
			} else {
				return errorSymbol(c, b.index)
			}
		case c == '+' || c == '-':
			if find == 0 || find == 8 {
				find |= 1
			} else {
				return errorSymbol(c, b.index)
			}
		case c == 'e' || c == 'E':
			if find&8 == 0 {
				find = 8
			} else {
				return errorSymbol(c, b.index)
			}
		default:
			if find&4 != 0 {
				return nil
			}
			return errorSymbol(c, b.index)
		}
	}
	if find&4 != 0 {
		return io.EOF
	}
	return errorEOF(b.index)
}

func (b *buffer) string() error {
	err := b.step()
	if err != nil {
		return errorEOF(b.index)
	}
	if !b.skip(quotes) {
		return errorEOF(b.index)
	}
	return nil
}

func (b *buffer) null() error {
	return b.word(_null)
}

func (b *buffer) true() error {
	return b.word(_true)
}

func (b *buffer) false() error {
	return b.word(_false)
}

func (b *buffer) word(word []byte) error {
	var c byte
	max := len(word)
	index := 0
	for ; b.index < b.length; b.index++ {
		c = b.data[b.index]
		if c != word[index] && c != (word[index]-32) {
			return errorSymbol(c, b.index)
		}
		index++
		if index >= max {
			break
		}
	}
	if index != max {
		return errorEOF(b.index)
	}
	return nil
}

func (b *buffer) step() error {
	if b.index+1 < b.length {
		b.index++
		return nil
	}
	return io.EOF
}
