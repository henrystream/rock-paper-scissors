package game

import "errors"

// DetermineWinner calculates the result of a match
func DetermineWinner(player1Choice, player2Choice string) (string, error) {
	if !isValidChoice(player1Choice) || !isValidChoice(player2Choice) {
		return "", errors.New("invalid choice")
	}

	if player1Choice == player2Choice {
		return "draw", nil
	}

	switch player1Choice {
	case "rock":
		if player2Choice == "scissors" {
			return "player1", nil
		}
		return "player2", nil
	case "paper":
		if player2Choice == "rock" {
			return "player1", nil
		}
		return "player2", nil
	case "scissors":
		if player2Choice == "paper" {
			return "player1", nil
		}
		return "player2", nil
	}
	return "", errors.New("unexpected choice combination")
}

func isValidChoice(choice string) bool {
	return choice == "rock" || choice == "paper" || choice == "scissors"
}
