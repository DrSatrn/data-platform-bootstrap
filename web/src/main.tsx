// This file boots the React application. Keeping startup explicit makes it
// easier to reason about rendering, global styles, and provider boundaries.
import React from "react";
import ReactDOM from "react-dom/client";

import { App } from "./app/App";
import "./styles/global.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
