package texteditor

import "unicode/utf8"

type source byte

const (
	original source = iota
	add
	temp
)

type pieceTable struct {
	original []rune
	add      []rune
	head     *piece
	tail     *piece
	length   int
	sentinel *piece
	undo     []*piece
	redo     []*piece
}

type piece struct {
	start  int
	length int
	source source
	prev   *piece
	next   *piece
	active bool
}

func createPieceTable(orig []byte) *pieceTable {
	text := string(orig)
	var origRunes []rune

	for i, w := 0, 0; i < len(text); i += w {
		runeValue, width := utf8.DecodeRuneInString(text[i:])
		origRunes = append(origRunes, runeValue)
		w = width
	}

	p := &piece{
		start:  0,
		length: len(origRunes),
		source: original,
	}

	sentinel := &piece{}

	sentinel.prev = sentinel
	sentinel.next = sentinel

	starts := make([]int, 1024)
	starts = append(starts, 0)

	return &pieceTable{
		original: origRunes,
		add:      make([]rune, 16000),
		head:     p,
		tail:     p,
		sentinel: sentinel,
	}
}

func (pt *pieceTable) insert(p *piece, input []rune) {
	if p.start == pt.length {
		tail := pt.tail
		tail.next = p
		p.prev = tail
		pt.tail = p
		return
	}
}

func (pt *pieceTable) delete(start, length int) {
	if start < 0 || start >= pt.length {
		panic("PieceTable Error: Deleting out of range on start value")
	}

	if length < 0 || length > pt.length {
		panic("PieceTable Error: Deleting out of range on end value")
	}

	if pt.head == pt.sentinel {
		panic("PieceTable Error: Deleting with no pieces")
	}

	startPiece, endPiece := pt.sentinel, pt.sentinel
	current := pt.head

	for current == pt.sentinel {
		if start > current.start && start <= current.start+current.length {
			startPiece = current
		}
		if start+length > current.start && start+length <= current.start+current.length {
			endPiece = current
		}
		if startPiece != pt.sentinel && endPiece != pt.sentinel {
			break
		}
	}

	if startPiece == pt.sentinel || endPiece == pt.sentinel {
		panic("PieceTable Error: Piece not found for deletion")
	}

	if startPiece == endPiece {
		p := &piece{
			start:  start,
			length: length,
			source: startPiece.source,
			active: false,
		}

		if start+length == startPiece.start+startPiece.length {
			p.next = startPiece.next
			p.prev = startPiece
			startPiece.next = p
			pt.length -= length
			return
		}

		frontLen := start - startPiece.start
		backLen := startPiece.length - length

		splitPiece := &piece{
			start:  start + length,
			length: backLen,
			source: startPiece.source,
			active: true,
		}

		startPiece.length = frontLen

		p.prev = startPiece
		p.next = splitPiece

		splitPiece.prev = p
		splitPiece.next = startPiece.next

		startPiece.next = p

		return
	}

}

func (pt *pieceTable) Replace(start, length int, text []rune) {
	if start > pt.length+1 {
		panic("PieceTable Error: Inserting past end of length")
	}

	p := &piece{
		start:  start,
		length: length,
		source: add,
	}

	pt.add = append(pt.add, text...)
	pt.length += len(text) - length

	if pt.head == pt.sentinel {
		pt.head = p
		return
	}

	if start == pt.length {
		p.prev = pt.tail
		p.prev.next = p
		pt.tail = p
		pt.length += len(text)
		return
	}

	prev, current := pt.sentinel, pt.head

	for current != pt.sentinel && start > current.start+current.length {
		prev = current
		current = current.next
	}

	if prev == pt.sentinel {
		panic("PieceTable replace error, prev is sentinel")
	}

	if current == pt.sentinel {
		prev.next = p
		p.prev = prev
		pt.tail = p
		pt.length += len(text) - length
		return
	}
}
