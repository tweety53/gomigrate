package version

import (
	"strconv"
	"testing"
)

func TestValidMigrationVersion(t *testing.T) {
	tests := []struct {
		v    string
		want bool
	}{
		{"", false},
		{"m12345_67890_q", false},
		{"m000000_000000_kek", true},
		{"m000000_000000_add_some_table_1", true},
		{"m000000_000000_kek+1", false},
		{"`m000000_000000_kek`", false},
		{"кириллица", false},
		{":)", false},
		{"some_file", false},
		{"123", false},
		{"m_1_2_a", false},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			if got := ValidMigrationVersion(tt.v); got != tt.want {
				t.Errorf("ValidMigrationVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
