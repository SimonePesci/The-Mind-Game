import React from "react";

interface PlayerHandProps {
  hand: number[];
  onPlayCard: (card: number) => void;
  onDiscardCard: (card: number) => void;
}

const PlayerHand: React.FC<PlayerHandProps> = ({
  hand,
  onPlayCard,
  onDiscardCard,
}) => {
  return (
    <div>
      <h2>Your Hand</h2>
      <div style={{ display: "flex", gap: "10px" }}>
        {hand
          .sort((a, b) => a - b)
          .map((card) => (
            <div key={card} style={{ textAlign: "center" }}>
              <p>{card}</p>
              <button onClick={() => onPlayCard(card)}>Play</button>
              <button onClick={() => onDiscardCard(card)}>Discard</button>
            </div>
          ))}
      </div>
    </div>
  );
};

export default PlayerHand;
