package game

import "testing"

func TestDetermineWinner(t *testing.T) {
	tests := []struct {
		p1Choice string
		p2Choice string
		want     string
		wantErr  bool
	}{
		{"rock", "scissors", "player1", false},
		{"paper", "rock", "player1", false},
		{"scissors", "paper", "player1", false},
		{"rock", "paper", "player2", false},
		{"rock", "rock", "draw", false},
		{"invalid", "rock", "", true},
	}

	for _, tt := range tests {
		got, err := DetermineWinner(tt.p1Choice, tt.p2Choice)
		if (err != nil) != tt.wantErr {
			t.Errorf("DetermineWinner(%q, %q) error = %v, wantErr %v", tt.p1Choice, tt.p2Choice, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("DetermineWinner(%q, %q) = %q, want %q", tt.p1Choice, tt.p2Choice, got, tt.want)
		}
	}
}
