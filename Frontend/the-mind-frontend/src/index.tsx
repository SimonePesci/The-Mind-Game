import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { WebSocketProvider } from "./context/WebSocketContext";

const container = document.getElementById("root");

if (container) {
  const root = ReactDOM.createRoot(container);
  root.render(
    <React.StrictMode>
      <WebSocketProvider>
        <App />
      </WebSocketProvider>
    </React.StrictMode>
  );
}
