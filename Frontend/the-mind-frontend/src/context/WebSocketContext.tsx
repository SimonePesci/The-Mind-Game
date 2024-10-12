import React, { createContext, useEffect, useState, ReactNode } from "react";

interface IWebSocketContext {
  ws: WebSocket | null;
}

interface IWebSocketProviderProps {
  children: ReactNode;
}

export const WebSocketContext = createContext<IWebSocketContext>({ ws: null });

export const WebSocketProvider: React.FC<IWebSocketProviderProps> = ({
  children,
}) => {
  const [ws, setWs] = useState<WebSocket | null>(null);

  useEffect(() => {
    const websocket = new WebSocket("ws://localhost:8080/ws");

    websocket.onopen = () => {
      console.log("WebSocket connection established");
    };

    websocket.onclose = () => {
      console.log("WebSocket connection closed");
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ ws }}>
      {children}
    </WebSocketContext.Provider>
  );
};
