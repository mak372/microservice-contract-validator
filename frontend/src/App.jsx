import { useState } from "react";
import ContractForm from "./components/ContractForm";
import TestRequest from "./components/TestRequest";
import ViolationsLog from "./components/ViolationsLog";
import "./App.css";

export default function App() {
  const [showViolations, setShowViolations] = useState(false);

  return (
    <div className="app">
      <header className="app-header">
        <h1>Contract Testing Proxy</h1>
        <button
          className="btn btn-outline"
          onClick={() => setShowViolations((v) => !v)}
        >
          {showViolations ? "Back" : "View Violations"}
        </button>
      </header>

      {showViolations ? (
        <ViolationsLog />
      ) : (
        <div className="panels">
          <ContractForm />
          <TestRequest />
        </div>
      )}
    </div>
  );
}
