package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/sleepy-day/sqline/texteditor"
)

func TestGapBufferInit(t *testing.T) {
	f, err := os.ReadFile("testfile.txt")
	if err != nil {
		t.Fatalf("Failed to open file: %s", err.Error())
	}

	gap := texteditor.CreateGapBuffer(f, 200)

	if len(gap.LineStarts()) != 8 {
		t.Fatalf("Failed to init LineStarts correctly: got %d", len(gap.LineStarts()))
	}

	buf := make([]rune, len(gap.Buf()))
	copy(buf, gap.Buf())

	for i := 0; i < 200; i++ {
		if buf[i] >= 0 {
			t.Fatalf("Found non empty char in buf")
		}
	}

	if len(buf) != 270 {
		t.Fatalf("Incorrect length in gap buffer: %d", len(buf))
	}
}

func TestMovementLimitedWithinChars(t *testing.T) {
	buf := []byte{}

	gap := texteditor.CreateGapBuffer(buf, 200)

	start, end := gap.GapStartAndEnd()

	gap.MoveLeft()
	gap.MoveLeft()
	gap.MoveLeft()
	gap.MoveLeft()

	startL, endL := gap.GapStartAndEnd()

	if start != startL || end != endL {
		t.Fatalf("Error on moving left: moved out of range.")
	}

	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()

	startR, endR := gap.GapStartAndEnd()

	if start != startR || end != endR {
		t.Fatalf("Error on moving right: moved out of range.")
	}
}

func TestMovement(t *testing.T) {
	f, err := os.ReadFile("testfile.txt")
	if err != nil {
		t.Fatalf("Failed to openn file: %s", err.Error())
	}

	gap := texteditor.CreateGapBuffer(f, 200)
	gapN := texteditor.CreateGapBuffer(f, 200)

	for range 5 {
		gap.MoveRight()
	}

	gapN.MoveNRight(5)

	buf := gap.Buf()
	bufN := gapN.Buf()

	for i := range buf {
		if buf[i] != bufN[i] {
			t.Fatal()
		}
	}

	for range 3 {
		gap.MoveLeft()
	}

	gapN.MoveNLeft(3)

	buf = gap.Buf()
	bufN = gapN.Buf()

	for i := range buf {
		if buf[i] != bufN[i] {
			t.Fatalf("Buffers do not match: %s is not equal to %s", string(buf[i]), string(bufN[i]))
		}
	}

	gap.MoveUp()
	gapN.MoveUp()

	buf = gap.Buf()
	bufN = gapN.Buf()

	for i := range buf {
		if buf[i] != bufN[i] {
			t.Fatalf("Buffers do not match: %s is not equal to %s", string(buf[i]), string(bufN[i]))
		}
	}

	gap.MoveNRight(15)
	gapN.MoveNRight(2)
	gapN.MoveDown()

	buf = gap.Buf()
	bufN = gapN.Buf()

	for i := range buf {
		if buf[i] != bufN[i] {
			t.Fatalf("Buffers do not match: %s is not equal to %s", string(buf[i]), string(bufN[i]))
		}
	}
}

func printBuf(buf []rune) {
	fmt.Println()

	for _, v := range buf {
		if v < 0 {
			fmt.Print("X")
		} else {
			fmt.Print(string(v))
		}
	}

	fmt.Println()
}

func TestGrow(t *testing.T) {
	f, err := os.ReadFile("testfile.txt")
	if err != nil {
		t.Fatalf("Failed to open file: %s", err.Error())
	}

	gap := texteditor.CreateGapBuffer(f, 200)
	gapCmp := texteditor.CreateGapBuffer(f, 200)

	gap.MoveNRight(20)
	gapCmp.MoveNRight(20)

	gap.Grow()

	buf := gap.Buf()
	bufCmp := gapCmp.Buf()

	start, _ := gap.GapStartAndEnd()
	for i := 0; i < start; i++ {
		if buf[i] != bufCmp[i] {
			t.Fatalf("Buffers do not match: %s is not equal to %s", string(buf[i]), string(bufCmp[i]))
		}
	}

	_, end := gap.GapStartAndEnd()
	_, endCmp := gapCmp.GapStartAndEnd()
	for i, j := end, endCmp; i < len(buf); i, j = i+1, j+1 {
		if buf[i] != bufCmp[j] {
			t.Fatalf("Buffers do not match: %s is not equal to %s", string(buf[i]), string(bufCmp[j]))
		}
	}
}

func TestGetLines(t *testing.T) {
	f, err := os.ReadFile("testfile.txt")
	if err != nil {
		t.Fatalf("Failed to open file: %s", err.Error())
	}

	gap := texteditor.CreateGapBuffer(f, 200)
	gapCmp := texteditor.CreateGapBuffer(f, 300)

	gap.MoveNRight(27)

	lines := gap.GetLines(1, 4)
	linesCmp := gapCmp.GetLines(1, 4)

	for i, v := range lines {
		for j, ch := range v {
			if ch != linesCmp[i][j] {
				t.Fatalf("Buffers do not match: %s is not equal to %s", string(ch), string(linesCmp[i][j]))
			}
		}
	}
}

/*

	gap.MoveNLeft(5)
	newBuf = make([]rune, len(gap.Buf()))
	copy(newBuf, gap.Buf())

	for i := 0; i < len(buf); i++ {
		if buf[i] != newBuf[i] {
			t.Fatalf("Error comparing buffers after MoveNLeft(): %s did not match %s", string(buf[i]), string(newBuf[i]))
		}
	}

	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()
	gap.MoveRight()

	copy(newBuf, gap.Buf())

	gap.MoveNLeft(5)

	gap.MoveNRight(5)

	for _, v := range gap.Buf() {
		if v < 0 {
			fmt.Print("X")
		} else {
			fmt.Print(string(v))
		}
	}
	fmt.Println()

	for _, v := range newBuf {
		if v < 0 {
			fmt.Print("X")
		} else {
			fmt.Print(string(v))
		}
	}

	for i, v := range gap.Buf() {
		if v != newBuf[i] {
			t.Fatalf("Error comparing buffers after MoveNRight(): %s did not match %s", string(v), string(newBuf[i]))
		}
	}
}
*/
