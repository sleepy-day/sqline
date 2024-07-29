package texteditor

var (
	tabSpaces = 4
)

type DocBuffer struct {
	bufs []GapBuffer
}

type GapBuffer struct {
	buf                        []rune
	chars                      int
	start, end                 int
	lineStarts                 []int
	lastMove                   int
	linePos                    int
	prevLineStart, prevLineEnd int
	tabsBehind                 int
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
		buf:   buf,
		chars: len(utfText),
		start: 0,
		end:   gapLength - 1,
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

func (gap *GapBuffer) Insert(ch rune) {
	if gap.start == gap.end {
		gap.Grow()
	}

	gap.prevX = -1

	gap.buf[gap.start] = ch
	gap.chars++
	gap.start++
}

func (gap *GapBuffer) Delete(backspace bool) {
	if backspace && gap.start == 0 {
		return
	}

	gap.chars--

	if backspace {
		gap.start--
		gap.buf[gap.start] = -1
		return
	}

	gap.buf[gap.end] = -1
	gap.end--
	return
}

func (gap *GapBuffer) Grow() {
	buf := make([]rune, len(gap.buf)+200)
	if gap.start > 0 {
		copy(buf[:gap.start], gap.buf[:gap.start])
	}
	copy(buf[gap.end:], gap.buf[gap.start:])

	for i := gap.start; i < gap.end; i++ {
		buf[i] = -1
	}

	gap.buf = buf
	gap.end += 200
}

func (gap *GapBuffer) MoveLeft() int {
	if gap.start == 0 {
		return 0
	}

	gap.prevX = -1

	gap.buf[gap.end-1] = rune(gap.buf[gap.start-1])
	gap.buf[gap.start-1] = rune(-1)

	gap.lastMove = -1

	gap.start--
	gap.end--

	gap.CalcLinePosition()
	return gap.tabsBehind
}

func (gap *GapBuffer) MoveRight() {
	if gap.start == len(gap.buf) || gap.end == len(gap.buf) {
		return
	}

	gap.prevX = -1

	gap.buf[gap.start] = rune(gap.buf[gap.end])
	gap.buf[gap.end] = rune(-1)

	gap.lastMove = +1

	gap.start++
	gap.end++

	gap.CalcLinePosition()
}

func (gap *GapBuffer) moveGapRight() {
	gap.start++
	gap.end++
}

func (gap *GapBuffer) MoveNLeft(n int) {
	if gap.start == 0 {
		return
	}

	if n > gap.start {
		n = gap.start
	}

	tmp := make([]rune, n)
	copy(tmp, gap.buf[gap.start-n:gap.start])
	copy(gap.buf[gap.start-n:gap.start], gap.buf[gap.end-n:gap.end])
	copy(gap.buf[gap.end-n:gap.end], tmp)

	gap.start -= n
	gap.end -= n
}

func (gap *GapBuffer) MoveNRight(n int) {
	if gap.end == len(gap.buf) {
		return
	}

	if gap.end+n > len(gap.buf) {
		n = gap.followingLinePos
	}

	tmp := make([]rune, n)
	copy(tmp, gap.buf[gap.end:gap.end+n])
	copy(gap.buf[gap.end:gap.end+n], gap.buf[gap.start:gap.start+n])
	copy(gap.buf[gap.start:gap.start+n], tmp)

	gap.start += n
	gap.end += n
}

func (gap *GapBuffer) MoveUp() int {
	gap.CalcLinePosition()
	gap.CalcPrevLineStart()

	if gap.prevX < 0 {
		gap.prevX = gap.linePos
	}

	if gap.prevLineStart <= 2 {
		gap.MoveNLeft(gap.prevLineStart)
	} else if gap.prevLineStart-gap.linePos > gap.prevX {
		gap.MoveNLeft(gap.prevLineStart - gap.prevX)
	} else {
		gap.MoveNLeft(gap.linePos + 1)
	}

	gap.CalcLinePosition()

	return gap.tabsBehind
}

func (gap *GapBuffer) MoveDown() int {
	gap.CalcLinePosition()
	gap.CalcNextLineStart()

	if gap.prevX < 0 {
		gap.prevX = gap.linePos
	}

	if gap.followingLinePos-gap.nextLinePos > gap.prevX {
		gap.MoveNRight(gap.nextLinePos + gap.prevX)
	} else if gap.followingLinePos+gap.start == gap.chars {
		gap.MoveNRight(gap.followingLinePos)
	} else {
		gap.MoveNRight(gap.followingLinePos - 1)
	}

	gap.CalcLinePosition()

	return gap.tabsBehind
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
	if gap.chars == 0 {
		return [][]rune{{}}
	}

	lines := make([][]rune, 0, end-start+1)

	linePositions := []LinePos{}

	startPos, endPos, line := 0, 0, 1
	lineStarted := false

	for i := 0; i < gap.start; i++ {
		if line < start && gap.buf[i] == '\n' {
			line++
			continue
		}

		if !lineStarted && gap.buf[i] == '\n' && line >= start && line <= end {
			if i < gap.start-2 && gap.buf[i+1] == '\n' {
				linePositions = append(linePositions, LinePos{
					start: i, end: i, splitLine: false,
				})
				line++
				continue
			}

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

	startPos = gap.end
	for i := gap.end; i < len(gap.buf); i++ {
		if line < start && gap.buf[i] == '\n' {
			line++
			continue
		}

		if !lineStarted && gap.buf[i] == '\n' && line >= start && line <= end {
			if i < gap.end-1 && gap.buf[i+1] == '\n' {
				linePositions = append(linePositions, LinePos{
					start: i, end: i, splitLine: false,
				})
				line++
				continue
			}

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
			start: startPos, end: len(gap.buf) - 1,
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
			lines = append(lines, []rune{})
			continue
		}

		lineBuf := append([]rune(nil), gap.buf[v.start:v.end]...)
		lines = append(lines, lineBuf)
	}

	if gap.buf[linePositions[len(linePositions)-1].end] == '\n' {
		lines = append(lines, []rune{})
	}

	return lines
}

func (gap *GapBuffer) Chars() int {
	return gap.chars
}

func (gap *GapBuffer) GapStartAndEnd() (int, int) {
	return gap.start, gap.end
}

func (gap *GapBuffer) LineStarts() []int {
	return gap.lineStarts
}

func (gap *GapBuffer) Buf() []rune {
	return gap.buf
}

func (gap *GapBuffer) NewLineBehindCursor() rune {
	if gap.start == 0 || gap.buf[gap.start-1] == '\n' {
		return 'E'
	}

	return ' '
}

func (gap *GapBuffer) CalcNextLineStart() {
	gap.nextLinePos = 0
	gap.followingLinePos = 0

	firstHit := false
	pos := 0
	for i := gap.end; i < len(gap.buf); i++ {
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
	for i := gap.start - 1; i >= 0; i-- {

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
	gap.linePos = 0
	gap.tabsBehind = 0
	if gap.start == 0 {
		return
	}

	for i := gap.start - 1; i >= 0; i-- {
		if gap.buf[i] == '\n' {
			break
		}
		if gap.buf[i] == '\t' {
			gap.tabsBehind++
		}
		gap.linePos++
	}
}

func (gap *GapBuffer) PeekBehind() rune {
	if gap.start == 0 {
		return -1
	}

	return gap.buf[gap.start-1]
}

func (gap *GapBuffer) PeekAhead() rune {
	if gap.end == len(gap.buf) {
		return -1
	}

	return gap.buf[gap.end]
}

func (gap *GapBuffer) CharAtCursor() rune {
	if gap.start == 0 {
		return -1
	}

	return gap.buf[gap.start-1]
}

func (gap *GapBuffer) LastMove() int {
	return gap.lastMove
}

func (gap *GapBuffer) PrevX() int {
	return gap.prevX
}
