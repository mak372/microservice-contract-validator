export default function HowToUse({ onClose }) {
  return (
    <div style={overlay}>
      <div style={modal}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "1.2rem" }}>
          <h2 style={{ margin: 0 }}>How to Use</h2>
          <button className="btn btn-outline" onClick={onClose}>✕ Close</button>
        </div>

        <p style={desc}>
          This app acts as a <strong>contract testing proxy</strong>. All traffic between services passes through it.
          It checks that every request and response matches the agreed contract and blocks anything that doesn't.
        </p>

        <Step number="1" title="Publish a Contract">
          Define what a request and response should look like for an endpoint.
          Fill in the method, endpoint path, upstream target URL, and the expected JSON schema for both request and response.
          <br /><br />
          <strong>Example:</strong>
          <Code>{`Method:   POST
Endpoint: /api/kyc/verify
Target:   http://localhost:8002

Request Schema:
{
  "customerId": "string",
  "fullName": "string",
  "documentNumber": "string"
}

Response Schema:
{
  "status": "string",
  "riskScore": "number"
}`}</Code>
        </Step>

        <Step number="2" title="Send a Request">
          Use the <strong>Test Request</strong> panel to send a request to the proxy.
          The proxy will validate your payload against the contract before forwarding it.
          <br /><br />
          If the request is missing a field or has the wrong type → <strong>blocked (400)</strong><br />
          If the upstream response breaks the contract → <strong>blocked (502)</strong><br />
          If everything matches → <strong>forwarded and returned normally</strong>
        </Step>

        <Step number="3" title="View Violations">
          Click <strong>View Violations</strong> in the header to see a log of every contract breach —
          which endpoint, which direction (request or response), and what fields failed.
        </Step>

        <Step number="4" title="View & Edit Contracts">
          Click <strong>View Contracts</strong> to see all published contracts.
          Click <strong>Edit</strong> on any contract to update its schema or target.
        </Step>

        <div style={{ marginTop: "1.5rem", padding: "0.75rem 1rem", background: "#1a1a2e", borderRadius: "6px", fontSize: "0.85rem", color: "#aaa" }}>
          <strong>Flow:</strong> Service A → Proxy (:8080) → Upstream service
        </div>
      </div>
    </div>
  );
}

function Step({ number, title, children }) {
  return (
    <div style={{ marginBottom: "1.2rem" }}>
      <div style={{ display: "flex", alignItems: "center", gap: "0.6rem", marginBottom: "0.4rem" }}>
        <span style={badge}>{number}</span>
        <strong>{title}</strong>
      </div>
      <div style={{ paddingLeft: "2rem", color: "#ccc", fontSize: "0.9rem", lineHeight: "1.6" }}>
        {children}
      </div>
    </div>
  );
}

function Code({ children }) {
  return (
    <pre style={{ background: "#111", padding: "0.75rem", borderRadius: "6px", fontSize: "0.8rem", marginTop: "0.5rem", overflowX: "auto" }}>
      {children}
    </pre>
  );
}

const overlay = {
  position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)",
  display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
};

const modal = {
  background: "#1e1e1e", borderRadius: "10px", padding: "2rem",
  width: "min(640px, 90vw)", maxHeight: "85vh", overflowY: "auto",
  boxShadow: "0 8px 32px rgba(0,0,0,0.5)",
};

const desc = { color: "#ccc", marginBottom: "1.5rem", lineHeight: "1.6" };

const badge = {
  background: "#4f8ef7", color: "#fff", borderRadius: "50%",
  width: "1.5rem", height: "1.5rem", display: "inline-flex",
  alignItems: "center", justifyContent: "center", fontSize: "0.8rem", fontWeight: "bold", flexShrink: 0,
};
