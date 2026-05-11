import { useEffect, useState } from "react";
import ContractForm from "./ContractForm";

const PROXY = import.meta.env.VITE_PROXY_URL || "http://localhost:8080";
const SESSION_START = new Date();

export default function ContractsList() {
  const [contracts, setContracts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(null);
  const [deleting, setDeleting] = useState(null);

  async function handleDelete(c) {
    if (!window.confirm(`Delete contract ${c.method} ${c.endpoint}?`)) return;
    setDeleting(c.method + c.endpoint);
    try {
      await fetch(`${PROXY}/contract/delete?method=${encodeURIComponent(c.method)}&endpoint=${encodeURIComponent(c.endpoint)}`, {
        method: "DELETE",
      });
      fetchContracts();
    } finally {
      setDeleting(null);
    }
  }

  async function fetchContracts() {
    setLoading(true);
    try {
      const res = await fetch(`${PROXY}/contracts`);
      const data = await res.json();
      const sessionContracts = data.filter(
        (c) => c.createdAt && new Date(c.createdAt) >= SESSION_START
      );
      setContracts(sessionContracts);
    } catch {
      setContracts([]);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { fetchContracts(); }, []);

  if (editing) {
    return (
      <div>
        <button className="btn btn-outline" style={{ marginBottom: "1rem" }} onClick={() => setEditing(null)}>
          ← Back to Contracts
        </button>
        <ContractForm
          initialContract={editing}
          onSaved={() => { setEditing(null); fetchContracts(); }}
        />
      </div>
    );
  }

  return (
    <div className="panel">
      <h2>Published Contracts</h2>
      {loading && <p>Loading...</p>}
      {!loading && contracts.length === 0 && <p>No contracts published this session.</p>}
      {!loading && contracts.length > 0 && (
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr>
              <th style={th}>Method</th>
              <th style={th}>Endpoint</th>
              <th style={th}>Target</th>
              <th style={th}></th>
            </tr>
          </thead>
          <tbody>
            {contracts.map((c) => (
              <tr key={c.method + c.endpoint}>
                <td style={td}><code>{c.method}</code></td>
                <td style={td}><code>{c.endpoint}</code></td>
                <td style={td}>{c.target}</td>
                <td style={td}>
                  <div style={{ display: "flex", gap: "0.5rem" }}>
                    <button className="btn btn-outline" onClick={() => setEditing(c)}>
                      Edit
                    </button>
                    <button
                      className="btn btn-outline"
                      style={{ color: "#e05c5c", borderColor: "#e05c5c" }}
                      disabled={deleting === c.method + c.endpoint}
                      onClick={() => handleDelete(c)}
                    >
                      {deleting === c.method + c.endpoint ? "Deleting..." : "Delete"}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

const th = { textAlign: "left", padding: "8px 12px", borderBottom: "1px solid #333" };
const td = { padding: "8px 12px", borderBottom: "1px solid #222" };
