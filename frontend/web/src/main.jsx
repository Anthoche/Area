/**
 * @file main.jsx
 * @description
 * Application entry point.
 *
 * Responsibilities:
 *  - Initializes the React application
 *  - Attaches the root component to the DOM
 *  - Enables React Strict Mode for development safeguards
 */

import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.jsx";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root")).render(
    <React.StrictMode>
        <App />
    </React.StrictMode>
);
