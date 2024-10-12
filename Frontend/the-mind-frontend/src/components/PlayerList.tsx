import React from "react";

interface PlayerListProps {
  players: string[];
}

const PlayerList: React.FC<PlayerListProps> = ({ players }) => {
  return (
    <div>
      <h2>Players in Room</h2>
      <ul>
        {players.map((playerId) => (
          <li key={playerId}>{playerId}</li>
        ))}
      </ul>
    </div>
  );
};

export default PlayerList;
