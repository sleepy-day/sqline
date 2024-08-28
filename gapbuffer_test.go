package main

import (
	"testing"

	"github.com/sleepy-day/sqline/util"
)

func TestGetLines(t *testing.T) {
	text := `Test Line One
Test Line Two
Test Line Three
Line Four Test
Testing Line Five
1
2
3
4
5`

	gap, _ := util.CreateGapBuffer([]byte(text), 200)

	lines := gap.GetLines(0, 99)

	if len(lines) != 10 {
		t.Fatalf("incorrect amount of lines returned, expected 10 got %d", len(lines))
	}

	if string(lines[0]) != "Test Line One\n" {
		t.Fatalf("line doesn't match %s", string(lines[0]))
	}
	if string(lines[1]) != "Test Line Two\n" {
		t.Fatalf("line doesn't match %s length %d", string(lines[1]), len(lines[1]))
	}
	if string(lines[2]) != "Test Line Three\n" {
		t.Fatalf("line doesn't match %s", string(lines[2]))
	}
	if string(lines[3]) != "Line Four Test\n" {
		t.Fatalf("line doesn't match %s", string(lines[3]))
	}
	if string(lines[4]) != "Testing Line Five\n" {
		t.Fatalf("line doesn't match %s", string(lines[4]))
	}
	if string(lines[5]) != "1\n" {
		t.Fatalf("line doesn't match %s", string(lines[5]))
	}
	if string(lines[6]) != "2\n" {
		t.Fatalf("line doesn't match %s", string(lines[6]))
	}
	if string(lines[7]) != "3\n" {
		t.Fatalf("line doesn't match %s", string(lines[7]))
	}
	if string(lines[8]) != "4\n" {
		t.Fatalf("line doesn't match %s", string(lines[8]))
	}
	if string(lines[9]) != "5" {
		t.Fatalf("line doesn't match %s", string(lines[9]))
	}

	lines = gap.GetLines(2, 6)

	if len(lines) != 5 {
		t.Fatalf("incorrect amount of lines returned, expected 5 got %d", len(lines))
	}

	if string(lines[0]) != "Test Line Three\n" {
		t.Fatalf("line doesn't match, got %s", string(lines[0]))
	}
	if string(lines[1]) != "Line Four Test\n" {
		t.Fatalf("line doesn't match %s", string(lines[1]))
	}
	if string(lines[2]) != "Testing Line Five\n" {
		t.Fatalf("line doesn't match %s", string(lines[2]))
	}
	if string(lines[3]) != "1\n" {
		t.Fatalf("line doesn't match %s", string(lines[3]))
	}
	if string(lines[4]) != "2\n" {
		t.Fatalf("line doesn't match %s", string(lines[4]))
	}
}

func TestGetLinesAfterInsert(t *testing.T) {
	text := `InsertHere->
And
Here->`

	gap, _ := util.CreateGapBuffer([]byte(text), 200)

	gap.Insert('X', util.Position{Line: 0, Col: 14})
	gap.Insert('X', util.Position{Line: 2, Col: 7})

	lines := gap.GetLines(0, 100)

	if string(lines[0]) != "InsertHere->X\n" {
		t.Fatalf("line doesn't match %s", string(lines[0]))
	}
	if string(lines[2]) != "Here->X" {
		t.Fatalf("line doesn't match %s", string(lines[2]))
	}
}

func TestGetTextInRange(t *testing.T) {
	text := `Test Line One
Test Line Two
Test Line Three
Line Four Test
Testing Line Five
1
2
3
4
5`

	expected := `ine Two
Test Line Three
Line Four Test
Testing Line Five
1
2
3
`

	gap, _ := util.CreateGapBuffer([]byte(text), 200)

	result, _ := gap.GetTextInRange(
		util.Position{Line: 1, Col: 6},
		util.Position{Line: 7, Col: 2},
	)

	if string(result) != expected {
		t.Fatalf("error in GetTextInRange, expected %s got %s", expected, string(result))
	}
}
