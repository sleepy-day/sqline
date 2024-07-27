package texteditor

import "fmt"

type DocBuffer struct {
	bufs []GapBuffer
}

type GapBuffer struct {
	buf                        []rune
	chars                      int
	gapStart, gapEnd           int
	lineStarts                 []int
	lastMove                   int
	linePos                    int
	prevLineStart, prevLineEnd int
	nextLinePos                int
	followingLinePos           int
	prevX                      int
}

type LinePos struct {
	start, end int
	splitLine  bool
}

func CreateGapBuffer(text []byte, gapLength int) *GapBuffer {
	var utfText []rune
	if len(text) > 0 && text[len(text)-1] == '\n' {
		utfText = []rune(string(text[:len(text)-1]))
	} else {
		utfText = []rune(string(text))
	}

	total := gapLength + len(utfText)
	buf := make([]rune, total)
	copy(buf[gapLength:], utfText)

	gap := &GapBuffer{
		buf:      buf,
		chars:    len(utfText),
		gapStart: 0,
		gapEnd:   gapLength,
	}

	i := 0
	for it := i; i < gapLength; it = i {
		buf[it] = -1
		i++
	}

	for it, j := i, 0; i < len(buf); it = i {
		if buf[it] == '\n' {
			gap.lineStarts = append(gap.lineStarts, j)
		}
		j++
		i++
	}

	return gap
}

func (gap *GapBuffer) GetLineCount() int {
	count := 1
	for _, v := range gap.buf {
		if v == '\n' {
			count++
		}
	}

	return count
}

func (gap *GapBuffer) GetLines(start, end int) [][]rune {
	lines := make([][]rune, 0, end-start+1)

	linePositions := []LinePos{}

	startPos, endPos, line := 0, 0, 1
	lineStarted := false

	for i := 0; i < gap.gapStart; i++ {
		if line < start && gap.buf[i] == '\n' {
			line++
			continue
		}

		if !lineStarted && gap.buf[i] == '\n' && line >= start && line <= end {
			linePositions = append(linePositions, LinePos{
				start: i, end: i, splitLine: false,
			})
			line++
			if line > end {
				break
			}
		} else if !lineStarted && line >= start && line <= end {
			startPos = i
			lineStarted = true
		} else if lineStarted && gap.buf[i] == '\n' {
			linePositions = append(linePositions, LinePos{
				start: startPos, end: i, splitLine: false,
			})
			lineStarted = false
			line++
			if line > end {
				break
			}
		}
		endPos = i
	}

	split := false
	if lineStarted {
		linePositions = append(linePositions, LinePos{
			start: startPos, end: endPos + 1, splitLine: true,
		})
		split = true
	}

	startPos = gap.gapEnd
	for i := gap.gapEnd; i < len(gap.buf); i++ {
		if line < start && gap.buf[i] == '\n' {
			line++
			continue
		}

		if !lineStarted && gap.buf[i] == '\n' && line >= start && line <= end {
			linePositions = append(linePositions, LinePos{
				start: i, end: i, splitLine: split,
			})
			line++
			split = false
			if line > end {
				break
			}
		} else if !lineStarted && line >= start && line <= end {
			startPos = i
			lineStarted = true
		} else if lineStarted && gap.buf[i] == '\n' {
			linePositions = append(linePositions, LinePos{
				start: startPos, end: i, splitLine: split,
			})
			lineStarted = false
			line++
			split = false
			if line > end {
				break
			}
		}
	}

	if lineStarted {
		linePositions = append(linePositions, LinePos{
			start: startPos, end: len(gap.buf),
		})
	}

	skip := false
	for i, v := range linePositions {
		if skip {
			skip = false
			continue
		}

		if v.splitLine && i+1 != len(linePositions) {
			lineBuf := append([]rune(nil), gap.buf[v.start:v.end]...)
			lineBuf = append(lineBuf, gap.buf[linePositions[i+1].start:linePositions[i+1].end]...)
			lines = append(lines, lineBuf)
			skip = true
			continue
		}

		if v.start == v.end {
			lines = append(lines, []rune{' '})
			continue
		}

		lineBuf := append([]rune(nil), gap.buf[v.start:v.end]...)
		lines = append(lines, lineBuf)
	}

	return lines
}

func (gap *GapBuffer) Chars() int {
	return gap.chars
}

func (gap *GapBuffer) GapStartAndEnd() (int, int) {
	return gap.gapStart, gap.gapEnd
}

func (gap *GapBuffer) LineStarts() []int {
	return gap.lineStarts
}

func (gap *GapBuffer) Buf() []rune {
	return gap.buf
}

func (gap *GapBuffer) Insert(ch rune) {
	if gap.gapStart == gap.gapEnd {
		gap.Grow()
	}

	gap.prevX = -1

	gap.buf[gap.gapStart] = ch
	gap.chars++
	gap.gapStart++
}

func (gap *GapBuffer) Grow() {
	buf := make([]rune, len(gap.buf)+200)
	if gap.gapStart > 0 {
		copy(buf[:gap.gapStart], gap.buf[:gap.gapStart])
	}
	copy(buf[gap.gapEnd:], gap.buf[gap.gapStart:])

	for i := gap.gapStart; i < gap.gapEnd; i++ {
		buf[i] = -1
	}

	gap.buf = buf
	gap.gapEnd += 200
}

func (gap *GapBuffer) MoveLeft() {
	if gap.gapStart == 0 {
		return
	}

	gap.prevX = -1

	gap.buf[gap.gapEnd-1] = rune(gap.buf[gap.gapStart-1])
	gap.buf[gap.gapStart-1] = rune(-1)

	gap.lastMove = -1

	gap.gapStart--
	gap.gapEnd--

	gap.CalcLinePosition()
}

func (gap *GapBuffer) MoveRight() {
	if gap.gapStart == len(gap.buf) {
		return
	}

	gap.prevX = -1

	gap.buf[gap.gapStart] = rune(gap.buf[gap.gapEnd])
	gap.buf[gap.gapEnd] = rune(-1)

	gap.lastMove = +1

	gap.gapStart++
	gap.gapEnd++

	gap.CalcLinePosition()
}

func (gap *GapBuffer) MoveNLeft(n int) {
	if gap.gapStart == 0 {
		return
	}

	if n > gap.gapStart {
		n = gap.gapStart
	}

	tmp := make([]rune, n)
	copy(tmp, gap.buf[gap.gapStart-n:gap.gapStart])
	copy(gap.buf[gap.gapStart-n:gap.gapStart], gap.buf[gap.gapEnd-n:gap.gapEnd])
	copy(gap.buf[gap.gapEnd-n:gap.gapEnd], tmp)

	gap.gapStart -= n
	gap.gapEnd -= n
}

func (gap *GapBuffer) MoveNRight(n int) {
	if gap.gapEnd == len(gap.buf) {
		return
	}

	if gap.gapEnd+n > len(gap.buf) {
		n = gap.followingLinePos
	}

	tmp := make([]rune, n)
	copy(tmp, gap.buf[gap.gapEnd:gap.gapEnd+n])
	copy(gap.buf[gap.gapEnd:gap.gapEnd+n], gap.buf[gap.gapStart:gap.gapStart+n])
	copy(gap.buf[gap.gapStart:gap.gapStart+n], tmp)

	gap.gapStart += n
	gap.gapEnd += n
}

// 1 is startGap - 2

func (gap *GapBuffer) NewLineBehindCursor() rune {
	if gap.gapStart == 0 || gap.buf[gap.gapStart-1] == '\n' {
		return 'E'
	}

	return ' '
}

func (gap *GapBuffer) MoveUp() {
	gap.CalcLinePosition()
	gap.CalcPrevLineStart()

	if gap.prevX < 0 {
		gap.prevX = gap.linePos
	}

	if gap.prevLineStart <= 2 {
		gap.MoveNLeft(gap.prevLineStart - 1)
	} else if gap.prevX > 0 && (gap.prevLineStart-gap.linePos) >= gap.prevX {
		fmt.Printf("%d", gap.prevLineStart)
		gap.MoveNLeft(gap.prevLineStart - gap.prevX)
		return
	} else if prevX > 0 && (gap.prevLineStart-gap.linePos) < gap.prevX {
		gap.MoveNLeft(gap.linePos + 1)
		return
	}

	if gap.prevLineStart-gap.linePos < gap.linePos {
		fmt.Print("/")
		gap.MoveNLeft(gap.linePos + 1)
		return
	}

	fmt.Print("\\")
	gap.MoveNLeft(gap.prevLineStart - gap.linePos)

	gap.CalcLinePosition()
	gap.CalcPrevLineStart()
}

// 1 is endGap

func (gap *GapBuffer) MoveDown() {
	gap.CalcLinePosition()
	gap.CalcNextLineStart()

	if gap.prevX < 0 {
		gap.prevX = gap.linePos
	}

	if gap.prevX >= 0 && (gap.followingLinePos-gap.nextLinePos) > gap.prevX {
		gap.MoveNRight(gap.nextLinePos + gap.prevX)
		return
	} else {
		gap.MoveNRight(gap.followingLinePos - 1)
		return
	}

	if gap.nextLinePos == 0 && gap.followingLinePos > 0 {
		gap.MoveNRight(gap.followingLinePos)
		return
	}

	if gap.nextLinePos+1 == gap.followingLinePos {
		gap.MoveNRight(gap.nextLinePos)
		return
	}

	if gap.followingLinePos+gap.gapEnd == len(gap.buf) {
		gap.MoveNRight(gap.followingLinePos)
		return
	}

	if gap.followingLinePos-gap.nextLinePos < gap.linePos {
		gap.MoveNRight(gap.followingLinePos - 1)
		return
	}

	gap.MoveNRight(gap.nextLinePos + gap.linePos)

	gap.CalcLinePosition()
	gap.CalcNextLineStart()
}

func (gap *GapBuffer) CalcNextLineStart() {
	gap.nextLinePos = 0
	gap.followingLinePos = 0

	firstHit := false
	pos := 0
	for i := gap.gapEnd; i < len(gap.buf); i++ {
		pos++
		if gap.buf[i] == '\n' {
			if !firstHit {
				gap.nextLinePos = pos
				firstHit = true
				continue
			}

			break
		}
	}

	gap.followingLinePos = pos
}

func (gap *GapBuffer) LinePositions() (int, int, int, int) {
	return gap.prevLineStart, gap.nextLinePos, gap.followingLinePos, gap.linePos
}

func (gap *GapBuffer) CalcPrevLineStart() {
	gap.prevLineStart = 0
	gap.prevLineEnd = 0

	firstHit := false
	for i := gap.gapStart - 1; i >= 0; i-- {
		if gap.buf[i] == '\n' {
			if !firstHit {
				firstHit = true
				gap.prevLineEnd = gap.prevLineStart
				gap.prevLineStart++
				continue
			}

			return
		}
		gap.prevLineStart++
	}
}

func (gap *GapBuffer) CalcLinePosition() {
	if gap.gapStart == 0 {
		gap.linePos = 0
	}

	pos := 0
	for i := gap.gapStart - 1; i >= 0; i-- {
		if gap.buf[i] == '\n' {
			break
		}
		pos++
	}

	gap.linePos = pos
}

func (gap *GapBuffer) CharAtCursor() rune {
	if gap.gapStart == 0 {
		return -1
	}

	return gap.buf[gap.gapStart-1]
}

func (gap *GapBuffer) LastMove() int {
	return gap.lastMove
}

func (gap *GapBuffer) PrevX() int {
	return gap.prevX
}
