package texteditor

import "fmt"

type DocBuffer struct {
	bufs []GapBuffer
}

type GapBuffer struct {
	buf              []rune
	chars            int
	gapStart, gapEnd int
	lineStarts       []int
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

	fmt.Printf("%+v\n", linePositions)

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

	gap.buf[gap.gapEnd-1] = rune(gap.buf[gap.gapStart-1])
	gap.buf[gap.gapStart-1] = rune(-1)

	gap.gapStart--
	gap.gapEnd--
}

func (gap *GapBuffer) MoveRight() {
	if gap.gapStart == gap.chars {
		return
	}

	gap.buf[gap.gapStart] = rune(gap.buf[gap.gapEnd])
	gap.buf[gap.gapEnd] = rune(-1)

	gap.gapStart++
	gap.gapEnd++
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

	if n+gap.gapStart > gap.chars {
		n = gap.chars - gap.gapStart
	}

	tmp := make([]rune, n)
	copy(tmp, gap.buf[gap.gapEnd:gap.gapEnd+n])
	copy(gap.buf[gap.gapEnd:gap.gapEnd+n], gap.buf[gap.gapStart:gap.gapStart+n])
	copy(gap.buf[gap.gapStart:gap.gapStart+n], tmp)

	gap.gapStart += n
	gap.gapEnd += n
}

func (gap *GapBuffer) MoveUp() {
	nlCount, steps, firstNl := 0, 0, 0
	for i := gap.gapStart; i >= 0; i-- {
		if gap.buf[i] == '\n' && nlCount == 0 {
			nlCount++
			firstNl = steps
		} else if gap.buf[i] == '\n' {
			nlCount++
			break
		}
		steps++
	}

	if nlCount == 0 {
		gap.MoveNLeft(gap.gapStart)
		return
	}

	if firstNl <= steps-firstNl {
		gap.MoveNLeft(steps - firstNl)
		return
	}

	if firstNl > steps-firstNl {
		gap.MoveNLeft(firstNl)
		return
	}
}

func (gap *GapBuffer) MoveDown() {
	nlCount, steps, firstNl := 0, 0, 0
	for i := gap.gapEnd; i < gap.gapEnd+gap.chars; i++ {
		if gap.buf[i] == '\n' && nlCount == 0 {
			nlCount++
			firstNl = steps
		} else if gap.buf[i] == '\n' {
			nlCount++
			break
		}
		steps++
	}

	linePos := 0
	for i := gap.gapStart; i >= 0; i-- {
		if gap.buf[i] == '\n' {
			break
		}

		linePos++
	}

	if nlCount == 0 {
		gap.MoveNRight(gap.chars)
		return
	}

	if firstNl <= steps-firstNl {
		gap.MoveNRight(firstNl + linePos)
		return
	}

	gap.MoveNRight(steps)
}
