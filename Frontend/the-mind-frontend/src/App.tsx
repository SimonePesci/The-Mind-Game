import React from "react";
import { useWebSocket } from "./hooks/useWebSocket";
import PlayerHand from "./components/PlayerHand";
import GameBoard from "./components/GameBoard";
import PlayerList from "./components/PlayerList";

const App: React.FC = () => {
  const { gameState, sendMessage, setGameState } = useWebSocket();
  const [connectionStatus, setConnectionStatus] =
    React.useState<string>("Connecting...");

  return (
    <div
      className="App"
      style={{ padding: "20px", fontFamily: "Arial, sans-serif" }}
    >
      <h1>The Mind Game</h1>
      <div style={{ marginBottom: "20px" }}>
        <p>
          <strong>Connection Status:</strong> {connectionStatus}
        </p>
        <p>
          <strong>Player ID:</strong> {gameState.playerId || "Loading..."}
        </p>
        <p>
          <strong>Room ID:</strong> {gameState.roomId || "Loading..."}
        </p>
        <p>
          <strong>Lives Left:</strong> {gameState.lives}
        </p>
        <p>
          <strong>Shurikens Left:</strong> {gameState.shurikens}
        </p>
      </div>
      <PlayerList players={gameState.players} />
      <GameBoard roundCards={gameState.roundCards} />
      <PlayerHand
        hand={gameState.hand}
        onPlayCard={(card) => {
          if (!gameState.playerId || !gameState.roomId) {
            console.error("Player ID or Room ID is missing.");
            return;
          }
          sendMessage("PLAY_CARD", {
            player_id: gameState.playerId,
            card_number: card,
          });
          // Optimistic Update
          setGameState((prevState) => ({
            ...prevState,
            hand: prevState.hand.filter((c) => c !== card),
          }));
        }}
        onDiscardCard={(card) => {
          if (!gameState.playerId || !gameState.roomId) {
            console.error("Player ID or Room ID is missing.");
            return;
          }
          sendMessage("DISCARD_CARD", {
            player_id: gameState.playerId,
            card_number: card,
          });
          // Optimistic Update
          setGameState((prevState) => ({
            ...prevState,
            hand: prevState.hand.filter((c) => c !== card),
          }));
        }}
      />
    </div>
  );
};

export default App;
