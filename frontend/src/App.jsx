import { useState } from "react";
import ContractForm from "./components/ContractForm";
import ContractsList from "./components/ContractsList";
import HowToUse from "./components/HowToUse";
import TestRequest from "./components/TestRequest";
import ViolationsLog from "./components/ViolationsLog";
import "./App.css";

export default function App() {
  const [view, setView] = useState("home"); // "home" | "violations" | "contracts"
  const [showHelp, setShowHelp] = useState(false);

  return (
    <div className="app">
      <header className="app-header">
        <h1>Contract Testing Proxy</h1>
        <div style={{ display: "flex", gap: "0.5rem" }}>
          <button className="btn btn-outline" onClick={() => setShowHelp(true)}>
            How to Use
          </button>
          <button className="btn btn-outline" onClick={() => setView(view === "contracts" ? "home" : "contracts")}>
            {view === "contracts" ? "Back" : "View Contracts"}
          </button>
          <button className="btn btn-outline" onClick={() => setView(view === "violations" ? "home" : "violations")}>
            {view === "violations" ? "Back" : "View Violations"}
          </button>
        </div>
      </header>

      {showHelp && <HowToUse onClose={() => setShowHelp(false)} />}

      {view === "violations" && <ViolationsLog />}
      {view === "contracts" && <ContractsList />}
      {view === "home" && (
        <div className="panels">
          <ContractForm />
          <TestRequest />
        </div>
      )}
    </div>
  );
}
