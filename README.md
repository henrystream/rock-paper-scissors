# Rock-Paper-Scissors Game Server

- A real-time multiplayer game server in Go using WebSockets.
- Global scoreboard updated after each match, also available via REST API

## Setup

1. Clone the repo:
   ```bash
   git clone https://github.com/henrystream/rock-paper-scissors.git
   cd rock-paper-scissor
   ```

## Run the Server

go run ./cmd/server

## Test with a WebSocket Client

Use wscat (install via npm install -g wscat) or a browser WebSocket client:

wscat -c ws://localhost:8080/ws

- Expected: {"event":"connected","playerId":"some-uuid"}

Send valid choice (after pairing with a second client):
{"event": "player_choice", "playerId": "<your-uuid>", "choice": "rock"}

**Open a Second Terminal**
Connect another client:

wscat -c ws://localhost:8080/ws

Both should receive "start_game", then proceed to send choices.

## Example messages:

- Connected: {"event": "connected", "playerId": "uuid"}
- Start Game: {"event": "start_game", "opponentId": "uuid"}
- Send Choice: {"event": "player_choice", "playerId": "uuid", "choice": "rock"}
- Game Result: {"event": "game_result", "winner": "uuid or 'draw'", "playerChoice": "rock", "opponentChoice": "scissors"}
- Scoreboard Update: `{"event": "scoreboard_update", "scoreboard": {"<playerId>": {"wins": 1, "losses": 0, "draws": 0}, ...}}`

## REST API

- **GET /scoreboard**: Returns the current scoreboard.
  - Example: `curl http://localhost:8080/scoreboard`
  - Response: `{"<playerId>": {"wins": 1, "losses": 0, "draws": 0}, ...}`
