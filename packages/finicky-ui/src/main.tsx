import { createRoot } from "react-dom/client";
import App from "./App";
import "./reset.css";
import "./app.css";

createRoot(document.getElementById("app")!).render(<App />);
