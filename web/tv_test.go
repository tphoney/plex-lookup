package web

import (
	_ "embed"
	"testing"

	"github.com/tphoney/plex-lookup/types"
)

func TestCleanTVSeries(t *testing.T) {
	original := []types.TVSeasonResult{
		{Number: 1},
		{Number: 2},
		{Number: 3},
		{Number: 4},
	}

	toRemove := []types.TVSeasonResult{
		{Number: 2},
		{Number: 4},
	}

	expected := []types.TVSeasonResult{
		{Number: 1},
		{Number: 3},
	}

	cleaned := cleanTVSeasons(original, toRemove)

	if len(cleaned) != len(expected) {
		t.Errorf("Expected %d cleaned TV series, but got %d", len(expected), len(cleaned))
	}

	for i := range cleaned {
		if cleaned[i].Number != expected[i].Number {
			t.Errorf("Expected cleaned TV series %v, but got %v", expected[i], cleaned[i])
		}
	}
}

func Test_discBeatsPlexResolution(t *testing.T) {
	type args struct {
		lowestResolution string
		format           []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "4K disc beats everything",
			args: args{
				lowestResolution: types.PlexResolution4K,
				format:           []string{types.Disk4K},
			},
			want: true,
		},
		{
			name: "Blu-ray disc beats 1080p",
			args: args{
				lowestResolution: types.PlexResolution1080,
				format:           []string{types.DiskBluray},
			},
			want: true,
		},
		{
			name: "Blu-ray disc is beaten by 4K",
			args: args{
				lowestResolution: types.PlexResolution4K,
				format:           []string{types.DiskBluray},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := discBeatsPlexResolution(tt.args.lowestResolution, tt.args.format); got != tt.want {
				t.Errorf("resolutionCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
