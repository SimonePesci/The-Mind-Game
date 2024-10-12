export interface GameState {
  playerId: string;
  roomId: string;
  hand: number[];
  roundCards: number[];
  lives: number;
  shurikens: number;
  players: string[];
}

export interface Message<T = any> {
  type: string;
  payload: T;
}

export interface WelcomePayload {
  message: string;
  player_id: string;
  room_id: string;
}

export interface NewCardPayload {
  card_number: number;
}

export interface NewCardsPayload {
  card_numbers: number[];
}

export interface PlayCardPayload {
  player_id: string;
  card_number: number;
}

export interface WrongCardPayload {
  player_id: string;
  card_number: number;
  position: number;
  lives_left: number;
}
