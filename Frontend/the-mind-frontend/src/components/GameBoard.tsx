import React from "react";

interface GameBoardProps {
  roundCards: number[];
}

const GameBoard: React.FC<GameBoardProps> = ({ roundCards }) => {
  return (
    <div>
      <h2>Cards Played This Round</h2>
      <div style={{ display: "flex", gap: "10px" }}>
        {roundCards
          .sort((a, b) => a - b)
          .map((card, index) => (
            <div
              key={index}
              style={{ border: "1px solid black", padding: "10px" }}
            >
              {card}
            </div>
          ))}
      </div>
    </div>
  );
};

export default GameBoard;
