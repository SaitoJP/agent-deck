package costs

import (
	"testing"
	"time"
)

func TestStartOfISOWeekUTC(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "monday stays in current week",
			now:  time.Date(2026, 5, 18, 15, 4, 5, 0, time.UTC),
			want: time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "tuesday rewinds to monday",
			now:  time.Date(2026, 5, 19, 1, 2, 3, 0, time.UTC),
			want: time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "sunday stays in same ISO week",
			now:  time.Date(2026, 5, 24, 23, 59, 59, 0, time.UTC),
			want: time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := startOfISOWeekUTC(tt.now); !got.Equal(tt.want) {
				t.Fatalf("startOfISOWeekUTC(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}
