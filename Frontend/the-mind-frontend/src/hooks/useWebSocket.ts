import { useContext, useEffect, useState } from "react";
import { WebSocketContext } from "../context/WebSocketContext";
import {
  GameState,
  Message,
  WelcomePayload,
  NewCardPayload,
  NewCardsPayload,
  PlayCardPayload,
  WrongCardPayload,
} from "../types";

export const useWebSocket = () => {
  const { ws } = useContext(WebSocketContext);
  const [gameState, setGameState] = useState<GameState>({
    playerId: "",
    roomId: "",
    hand: [],
    roundCards: [],
    lives: 3,
    shurikens: 1,
    players: [],
  });

  useEffect(() => {
    if (!ws) return;

    ws.onmessage = (event: MessageEvent) => {
      const message: Message = JSON.parse(event.data);
      console.log("Received message:", message);

      handleServerMessage(message);
    };
  }, [ws]);

  const handleServerMessage = (message: Message) => {
    switch (message.type) {
      case "WELCOME":
        handleWelcomeMessage(message.payload as WelcomePayload);
        break;
      case "NEW_CARD":
        handleNewCardMessage(message.payload as NewCardPayload);
        break;
      case "NEW_CARDS":
        handleNewCardsMessage(message.payload);
        break;
      case "CARD_PLAYED":
        handleCardPlayedMessage(message.payload);
        break;
      case "WRONG_CARD":
        handleWrongCardMessage(message.payload);
        break;
      // Add more cases as needed
      default:
        console.warn("Unknown message type:", message.type);
    }
  };

  const handleWelcomeMessage = (payload: WelcomePayload) => {
    console.log(payload);

    setGameState((prevState) => ({
      ...prevState,
      playerId: payload.player_id,
      roomId: payload.room_id,
    }));
  };

  const handleNewCardMessage = (payload: NewCardPayload) => {
    console.log(payload);
    setGameState((prevState) => ({
      ...prevState,
      hand: [...prevState.hand, payload.card_number],
    }));
  };

  const handleNewCardsMessage = (payload: NewCardsPayload) => {
    console.log(payload);

    setGameState((prevState) => ({
      ...prevState,
      hand: [...prevState.hand, ...payload.card_numbers],
    }));
  };

  const handleCardPlayedMessage = (payload: PlayCardPayload) => {
    setGameState((prevState) => ({
      ...prevState,
      roundCards: [...prevState.roundCards, payload.card_number],
    }));
  };

  const handleWrongCardMessage = (payload: WrongCardPayload) => {
    setGameState((prevState) => ({
      ...prevState,
      lives: payload.lives_left,
    }));

    if (payload.lives_left <= 0) {
      alert("Game Over! All lives lost.");
    } else {
      alert(
        `Player ${payload.player_id} played a wrong card: ${payload.card_number}`
      );
    }
  };

  const sendMessage = (type: string, payload: object) => {
    if (!ws) return;

    const message: Message = {
      type: type,
      payload: payload,
    };

    console.log(message);

    ws.send(JSON.stringify(message));
  };

  return {
    gameState,
    sendMessage,
    setGameState, // Expose setGameState to update state locally
  };
};
