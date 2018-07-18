package youtubecontent

import "testing"

func TestComplexParseInt(t *testing.T) {
	tables := []struct {
		num     string
		base    int
		bitSize int
		output  int64
	}{
		{"", 10, 0, 0},
		{"1", 10, 0, 1},
		{"-1", 10, 0, -1},
		{"5.3", 10, 0, 0},
	}

	for _, table := range tables {
		total := ComplexParseInt(table.num, table.base, table.bitSize)
		if total != table.output {
			t.Errorf("ComplexParseInt of (%s, %d, %d) was incorrect, got: %d, want: %d.", table.num, table.base, table.bitSize, total, table.output)
		}
	}
}

func TestParseAPIDurationResponse(t *testing.T) {
	tables := []struct {
		durationString string
		output         int64
	}{
		{"", 0},
		{"PTM0S", 0},
		{"PT1MS", 60},
		{"PT0M1S", 1},
		{"PT59M1S", 59*60 + 1},
		{"PT1H0M0S", 3600},
		{"P2DT1H0M1S", ((2*24)+1)*3600 + 1},
	}

	for _, table := range tables {
		total := ParseAPIDurationResponse(table.durationString)
		if total != table.output {
			t.Errorf("ParseAPIDurationResponse of (%s) was incorrect, got: %d, want: %d.", table.durationString, total, table.output)
		}
	}
}
