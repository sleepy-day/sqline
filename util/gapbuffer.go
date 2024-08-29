package util

import (
	"errors"
	"unicode/utf8"
)

var (
	tabSpaces = 4

	ErrNonUTF8Text  = errors.New("text contains non unicode characters")
	ErrOutOfBounds  = errors.New("offset out of bounds")
	ErrInvalidRange = errors.New("end position is greater than start position")
)

type DocBuffer struct {
	bufs []GapBuffer
}

type GapBuffer struct {
	buf      []rune
	gapStart int
	gapLen   int
	lines    int
}

type Pos struct {
	Line int
	Col  int
}

type BufRange struct {
	Start Pos
	End   Pos
}

func CreateGapBuffer(text []byte, gapLength int) (*GapBuffer, error) {
	if len(text) > 0 && !utf8.Valid(text) {
		return &GapBuffer{buf: make([]rune, gapLength), gapStart: 0, gapLen: gapLength}, ErrNonUTF8Text
	}

	chars := []rune(string(text))
	lines := 0
	for _, ch := range chars {
		if ch == '\n' {
			lines++
		}
	}

	buf := make([]rune, gapLength+len(chars))
	copy(buf[:len(chars)], chars)

	for i := len(chars); i < len(buf); i++ {
		buf[i] = 0
	}

	gap := &GapBuffer{
		buf:      buf,
		gapStart: len(chars),
		gapLen:   len(buf) - len(chars),
		lines:    lines,
	}

	return gap, nil
}

func (gap *GapBuffer) Insert(ch rune, pos Pos) error {
	if !utf8.ValidRune(ch) {
		return ErrNonUTF8Text
	}

	if gap.gapLen == 0 {
		gap.gapStart = len(gap.buf)
		gap.Grow()
		gap.gapLen = len(gap.buf) - gap.gapStart
	}

	offset, err := gap.FindOffset(pos)
	if err != nil {
		return err
	}

	gap.ShiftGap(offset)
	gap.buf[gap.gapStart] = ch
	gap.gapStart++
	gap.gapLen--

	if ch == '\n' {
		gap.lines++
	}

	return nil
}

func (gap *GapBuffer) Start() int {
	return gap.gapStart
}

func (gap *GapBuffer) ShiftGap(offset int) {
	if gap.gapLen == 0 {
		return
	}

	if offset < gap.gapStart {
		for i := gap.gapStart - 1; i >= offset; i-- {
			gap.buf[i+gap.gapLen] = gap.buf[i]
			gap.buf[i] = 0
		}

		gap.gapStart = offset
	} else if offset > gap.gapStart {
		for i := gap.gapStart + gap.gapLen; i < offset; i++ {
			gap.buf[i-gap.gapLen] = gap.buf[i]
			gap.buf[i] = 0
		}

		gap.gapStart = offset - gap.gapLen
	}

}

func (gap *GapBuffer) Delete(backward bool) {
	if backward && gap.gapStart == 0 {
		return
	} else if !backward && gap.gapStart+gap.gapLen >= len(gap.buf) {
		return
	}

	if backward {
		if gap.buf[gap.gapStart-1] == '\n' {
			gap.lines--
		}
		gap.buf[gap.gapStart-1] = 0
		gap.gapStart--
		gap.gapLen++
		return
	}

	if gap.buf[gap.gapStart+gap.gapLen] == '\n' {
		gap.lines--
	}

	gap.buf[gap.gapStart+gap.gapLen] = 0
	gap.gapLen++
	return
}

func (gap *GapBuffer) DeleteRange(rng BufRange) error {
	startOffset, err := gap.FindOffset(rng.Start)
	if err != nil {
		return err
	}

	if startOffset > 0 {
		for i := gap.gapStart; i < gap.gapStart+startOffset; i++ {
			if gap.buf[i] == '\n' {
				gap.lines--
			}
		}
	} else if startOffset < 0 {
		for i := gap.gapStart + startOffset; i < gap.gapStart; i++ {
			if gap.buf[i] == '\n' {
				gap.lines--
			}
		}
	}

	gap.ShiftGap(startOffset)

	endOffset, err := gap.FindOffset(rng.End)
	if err != nil {
		return err
	}

	for i := gap.gapLen; i < endOffset-gap.gapStart+1; i++ {
		if gap.buf[i] == '\n' {
			gap.lines--
		}
	}

	gap.gapLen = endOffset - gap.gapStart + 1
	return nil
}

func (gap *GapBuffer) FindOffset(pos Pos) (int, error) {
	startBuf := gap.buf[:gap.gapStart]
	line := 0
	col := 0

	for i := 0; i < len(startBuf); i++ {
		if line == pos.Line && col == pos.Col {
			return i, nil
		}

		if line == pos.Line && col < pos.Col-1 && i < len(startBuf)-1 && startBuf[i+1] == '\n' {
			return i + 1, nil
		}

		if startBuf[i] == '\n' {
			line += 1
			col = 0
		} else {
			col++
		}
	}

	if line == pos.Line && col == pos.Col {
		offset := gap.gapStart + gap.gapLen
		return offset, nil
	}

	endBuf := gap.buf[gap.gapStart+gap.gapLen:]
	for i := 0; i < len(endBuf); i++ {
		if line == pos.Line && col == pos.Col {
			return gap.gapStart + gap.gapLen + i, nil
		}

		if line == pos.Line && col < pos.Col-1 && i < len(endBuf)-1 && endBuf[i+1] == '\n' {
			return gap.gapStart + gap.gapLen + i + 1, nil
		}

		if i == len(endBuf)-1 {
			return gap.gapStart + gap.gapLen + i + 1, nil
		}

		if endBuf[i] == '\n' {
			line += 1
			col = 0
		} else {
			col++
		}
	}

	if line == pos.Line && col == pos.Col {
		offset := len(gap.buf)
		return offset, nil
	}

	return 0, ErrOutOfBounds
}

func (gap *GapBuffer) Grow() {
	buf := make([]rune, len(gap.buf)+8000)
	copy(buf[:gap.gapStart], gap.buf[:gap.gapStart])

	for i := gap.gapStart; i < len(gap.buf); i++ {
		buf[i] = 0
	}

	gap.buf = buf
}

func (gap *GapBuffer) Lines() int {
	return gap.lines
}

func (gap *GapBuffer) GetTextInRange(start, end Pos) ([]rune, error) {
	if start.Line > end.Line || (start.Line == end.Line && end.Col < start.Col) {
		return nil, ErrInvalidRange
	}

	lines := gap.GetLines(start.Line, end.Line)
	if len(lines) != end.Line-start.Line+1 {
		panic("invalid amount of lines returned")
	}

	if start.Line == end.Line {
		return lines[0][start.Col:end.Col], nil
	}

	lineCount := end.Line - start.Line + 1
	var text []rune
	for i, v := range lines {
		switch {
		case i == 0:
			text = append(text, v[start.Col:]...)
		case i == lineCount:
			text = append(text, v[:end.Col]...)
		default:
			text = append(text, v...)
		}
	}

	return text, nil
}

func (gap *GapBuffer) GetLines(start, end int) [][]rune {
	lines := [][]rune{}

	startBuf := gap.buf[:gap.gapStart]
	endBuf := gap.buf[gap.gapStart+gap.gapLen:]

	line := 0
	startPos := 0
	leftOverStartPos := -1

	for i := 0; i < len(startBuf); i++ {
		switch {
		case i == len(startBuf)-1 && line >= start && line <= end && len(endBuf) == 0:
			lines = append(lines, append([]rune(nil), startBuf[startPos:i+1]...))
			if startBuf[i] == '\n' {
				lines = append(lines, []rune{})
			}

			return lines
		case i == len(startBuf)-1 && line >= start && line <= end:
			if startBuf[i] == '\n' {
				lines = append(lines, append([]rune(nil), startBuf[startPos:]...))
			} else {
				leftOverStartPos = startPos
			}
		case line < start:
			if startBuf[i] == '\n' {
				startPos = i + 1
				line++
				continue
			}
			continue
		case line >= start && line <= end:
			if startBuf[i] == '\n' {
				lines = append(lines, append([]rune(nil), startBuf[startPos:i+1]...))
				startPos = i + 1
				line++
			}
		case line > end:
			return lines
		}
	}

	startPos = 0
	for i := 0; i < len(endBuf); i++ {
		switch {
		case i == len(endBuf)-1 && line >= start && line <= end:
			if leftOverStartPos != -1 {
				tmp := append([]rune(nil), startBuf[leftOverStartPos:]...)
				tmp = append(tmp, endBuf[startPos:i+1]...)
				lines = append(lines, tmp)

				if endBuf[i] == '\n' {
					lines = append(lines, []rune{})
				}
			} else {
				lines = append(lines, append([]rune(nil), endBuf[startPos:i+1]...))
			}
		case line < start:
			if start == 0 && end == 3 {
				panic(2)
			}
			if endBuf[i] == '\n' {
				startPos = i + 1
				line++
			}
		case line >= start && line <= end:
			if start == 0 && end == 3 {
				panic(3)
			}
			if endBuf[i] == '\n' && leftOverStartPos != -1 {
				tmp := append([]rune(nil), startBuf[leftOverStartPos:]...)
				tmp = append(tmp, endBuf[:i+1]...)
				lines = append(lines, tmp)
				leftOverStartPos = -1
			} else if endBuf[i] == '\n' {
				lines = append(lines, append([]rune(nil), endBuf[startPos:i+1]...))
			}

			if endBuf[i] == '\n' {
				startPos = i + 1
				line++
			}
		case line > end:
			if start == 0 && end == 3 {
				panic(4)
			}
			return lines
		}
	}

	return lines
}

func (gap *GapBuffer) PeekBehind() rune {
	if gap.gapStart > 0 {
		return gap.buf[gap.gapStart-1]
	}

	return 0
}

func (gap *GapBuffer) TabsBehind() int {
	buf := gap.buf[:gap.gapStart]
	count := 0

	for i := len(buf) - 1; i >= 0; i-- {
		if buf[i] == '\n' {
			return count
		} else if buf[i] == '\t' {
			count++
		}
	}

	return count
}

func (gap *GapBuffer) Buf() []rune {
	return gap.buf
}
