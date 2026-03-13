import { useEffect, useRef, useState } from "react";
import {
  Server,
  Play,
  Square,
  Save,
  RotateCw,
  Plus,
  Trash2,
  X,
  Globe,
  Cable,
  Network,
} from "lucide-react";
import toast from "react-hot-toast";
import { colors } from "./theme";

interface DnsConfigData {
  addr: string;
  port: number;
  network: string;
  cache_ttl: number;
  running: boolean;
}

interface Forwarder {
  id: number;
  addr: string;
  port: number;
  network: string;
}

export default function DnsServer() {
  const [addr, setAddr] = useState("");
  const [port, setPort] = useState(0);
  const [network, setNetwork] = useState("udp");
  const [cacheTTL, setCacheTTL] = useState(0);
  const [running, setRunning] = useState(false);
  const [hasConfig, setHasConfig] = useState(false);
  const [loading, setLoading] = useState(false);
  const [needsRestart, setNeedsRestart] = useState(false);

  const [forwarders, setForwarders] = useState<Forwarder[]>([]);
  const [fwAddr, setFwAddr] = useState("");
  const [fwPort, setFwPort] = useState(53);
  const [fwNetwork, setFwNetwork] = useState("udp");
  const [showFwForm, setShowFwForm] = useState(false);

  const savedConfig = useRef<{
    addr: string;
    port: number;
    network: string;
    cache_ttl: number;
  } | null>(null);

  useEffect(() => {
    fetchConfig();
    fetchForwarders();
  }, []);

  function fetchConfig() {
    fetch("/api/dns/config", { credentials: "same-origin" })
      .then((r) => {
        if (!r.ok) {
          setHasConfig(false);
          return null;
        }
        return r.json();
      })
      .then((data: DnsConfigData | null) => {
        if (data) {
          setAddr(data.addr);
          setPort(data.port);
          setNetwork(data.network);
          setCacheTTL(data.cache_ttl);
          setRunning(data.running);
          setHasConfig(true);
          savedConfig.current = {
            addr: data.addr,
            port: data.port,
            network: data.network,
            cache_ttl: data.cache_ttl,
          };
        }
      })
      .catch(() => {});
  }

  function fetchForwarders() {
    fetch("/api/dns/forwarders", { credentials: "same-origin" })
      .then((r) => r.json())
      .then((data) => setForwarders(data || []))
      .catch(() => {});
  }

  async function handleSave() {
    setLoading(true);
    try {
      const res = await fetch("/api/dns/config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "same-origin",
        body: JSON.stringify({ addr, port, network, cache_ttl: cacheTTL }),
      });
      const data = await res.json();
      if (!res.ok) {
        toast.error(data.error || "Failed to save configuration");
        return;
      }
      setRunning(data.running);
      setHasConfig(true);
      toast.success("Configuration saved");

      const prev = savedConfig.current;
      if (
        data.running &&
        prev &&
        (prev.addr !== addr ||
          prev.port !== port ||
          prev.network !== network ||
          prev.cache_ttl !== cacheTTL)
      ) {
        setNeedsRestart(true);
      }

      savedConfig.current = { addr, port, network, cache_ttl: cacheTTL };
    } catch {
      toast.error("Network error");
    } finally {
      setLoading(false);
    }
  }

  async function handleStartStop() {
    setLoading(true);
    const wasStopped = !running;
    const endpoint = running ? "/api/dns/stop" : "/api/dns/start";
    try {
      const res = await fetch(endpoint, {
        method: "POST",
        credentials: "same-origin",
      });
      const data = await res.json();
      if (!res.ok) {
        toast.error(data.error || "Operation failed");
        return;
      }
      setRunning(data.running);
      setNeedsRestart(false);
      toast.success(wasStopped ? "DNS server started" : "DNS server stopped");
    } catch {
      toast.error("Network error");
    } finally {
      setLoading(false);
    }
  }

  async function handleRestart() {
    setLoading(true);
    try {
      const res = await fetch("/api/dns/restart", {
        method: "POST",
        credentials: "same-origin",
      });
      const data = await res.json();
      if (!res.ok) {
        toast.error(data.error || "Failed to restart");
        return;
      }
      setRunning(data.running);
      setNeedsRestart(false);
      toast.success("DNS server restarted");
    } catch {
      toast.error("Network error");
    } finally {
      setLoading(false);
    }
  }

  async function handleAddForwarder() {
    if (!fwAddr || !fwPort) return;
    try {
      const res = await fetch("/api/dns/forwarders", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "same-origin",
        body: JSON.stringify({ addr: fwAddr, port: fwPort, network: fwNetwork }),
      });
      const data = await res.json();
      if (!res.ok) {
        toast.error(data.error || "Failed to add forwarder");
        return;
      }
      setForwarders((prev) => [...prev, data]);
      setFwAddr("");
      setFwPort(53);
      setFwNetwork("udp");
      toast.success("Forwarder added");
    } catch {
      toast.error("Network error");
    }
  }

  async function handleDeleteForwarder(id: number) {
    try {
      const res = await fetch(`/api/dns/forwarders/${id}`, {
        method: "DELETE",
        credentials: "same-origin",
      });
      if (!res.ok) {
        const data = await res.json();
        toast.error(data.error || "Failed to delete forwarder");
        return;
      }
      setForwarders((prev) => prev.filter((f) => f.id !== id));
      toast.success("Forwarder removed");
    } catch {
      toast.error("Network error");
    }
  }

  return (
    <div
      style={{
        padding: "2rem",
        display: "flex",
        flexDirection: "column",
        gap: "1.5rem",
      }}>
      {/* DNS Server Config */}
      <div style={cardStyle}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "1.25rem",
            paddingBottom: "0.75rem",
            borderBottom: `1px solid ${colors.border}`,
          }}>
          <h2
            style={{
              margin: 0,
              fontSize: "1rem",
              fontWeight: 600,
              display: "flex",
              alignItems: "center",
              gap: "0.5rem",
            }}>
            <Server size={18} />
            DNS Server
          </h2>
          {hasConfig && (
            <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
              <span
                style={{
                  display: "inline-block",
                  width: 10,
                  height: 10,
                  borderRadius: "50%",
                  background: running ? "#16a34a" : "#dc2626",
                  animation: running ? "pulse-dot 1.5s ease-in-out infinite" : "none",
                }}
              />
              <button
                onClick={handleStartStop}
                disabled={loading}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "0.3rem",
                  padding: "0.3rem 0.8rem",
                  background: running
                    ? "rgba(220,38,38,0.08)"
                    : "rgba(22,163,106,0.08)",
                  color: running ? "#dc2626" : "#16a34a",
                  border: `1px solid ${running ? "#dc2626" : "#16a34a"}`,
                  borderRadius: "9999px",
                  cursor: loading ? "not-allowed" : "pointer",
                  fontWeight: 600,
                  fontSize: "0.75rem",
                  fontFamily: "inherit",
                }}>
                {running ? <Square size={12} /> : <Play size={12} />}
                {running ? "Stop" : "Start"}
              </button>
            </div>
          )}
        </div>

        {needsRestart && (
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              padding: "0.6rem 0.8rem",
              marginBottom: "0.75rem",
              background: "rgba(245,158,11,0.08)",
              border: "1px solid #f59e0b",
              borderRadius: "6px",
              fontSize: "0.8rem",
              color: "#92400e",
            }}>
            <span>
              Configuration changed. Restart the DNS server to apply.
            </span>
            <button
              onClick={handleRestart}
              disabled={loading}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.3rem",
                padding: "0.3rem 0.7rem",
                background: "#f59e0b",
                color: "#fff",
                border: "none",
                borderRadius: "4px",
                cursor: loading ? "not-allowed" : "pointer",
                fontWeight: 600,
                fontSize: "0.75rem",
                fontFamily: "inherit",
                whiteSpace: "nowrap",
              }}>
              <RotateCw size={12} />
              Restart
            </button>
          </div>
        )}

        <div
          style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
          <div style={{ display: "flex", gap: "0.75rem" }}>
            <div style={{ flex: 1 }}>
              <label style={labelStyle}>Address</label>
              <input
                type="text"
                placeholder="0.0.0.0"
                value={addr}
                onChange={(e) => setAddr(e.target.value)}
                style={inputStyle}
              />
            </div>
            <div style={{ width: "120px" }}>
              <label style={labelStyle}>Port</label>
              <input
                type="number"
                min={1}
                max={65535}
                placeholder="5355"
                value={port || ""}
                onChange={(e) => setPort(Number(e.target.value))}
                style={inputStyle}
              />
            </div>
          </div>

          <div>
            <label style={labelStyle}>Protocol</label>
            <select
              value={network}
              onChange={(e) => setNetwork(e.target.value)}
              style={inputStyle}>
              <option value="udp">UDP</option>
              <option value="tcp">TCP</option>
            </select>
          </div>

          <div>
            <label style={labelStyle}>Cache TTL (seconds)</label>
            <input
              type="number"
              min={0}
              placeholder="14400"
              value={cacheTTL}
              onChange={(e) => setCacheTTL(Number(e.target.value))}
              style={inputStyle}
            />
          </div>

          <div style={{ display: "flex", justifyContent: "flex-end", marginTop: "0.75rem", paddingTop: "0.75rem" }}>
            <button
              onClick={handleSave}
              disabled={loading || !addr || !port}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.3rem",
                padding: "0.6rem 1.2rem",
                background:
                  !loading && addr && port ? colors.primary : colors.border,
                color: !loading && addr && port ? "#fff" : "#9ca3af",
                border: "none",
                borderRadius: "6px",
                cursor: !loading && addr && port ? "pointer" : "not-allowed",
                fontWeight: 600,
                fontSize: "0.85rem",
                fontFamily: "inherit",
              }}>
              <Save size={14} />
              Save
            </button>
          </div>
        </div>
      </div>

      {/* Upstream Forwarders */}
      <div style={cardStyle}>
        <h2 style={cardTitleStyle}>
          <Server size={18} />
          Upstream Forwarders
        </h2>

        <p
          style={{
            margin: "0 0 0.75rem",
            fontSize: "0.8rem",
            color: "#64748b",
            lineHeight: 1.5,
          }}>
          DNS queries are forwarded using round-robin with failover. If the
          primary server fails, the next one in the list is tried.
        </p>

        <div style={{ overflow: "auto" }}>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: "0.85rem",
            }}>
            <thead>
              <tr
                style={{
                  borderBottom: `1px solid ${colors.border}`,
                  textAlign: "left",
                }}>
                <th style={thStyle}><span style={thInnerStyle}><Globe size={12} /> Address</span></th>
                <th style={thStyle}><span style={thInnerStyle}><Cable size={12} /> Port</span></th>
                <th style={thStyle}><span style={thInnerStyle}><Network size={12} /> Protocol</span></th>
                <th style={{ ...thStyle, width: "1%" }}></th>
              </tr>
            </thead>
            <tbody>
              {forwarders.map((f) => (
                <tr
                  key={f.id}
                  style={{ borderBottom: `1px solid ${colors.border}` }}>
                  <td style={tdStyle}><span style={cellIconStyle}><Globe size={13} /> {f.addr}</span></td>
                  <td style={tdStyle}><span style={cellIconStyle}><Cable size={13} /> <span style={{ fontSize: "0.7rem", color: "#64748b", background: "#f1f5f9", border: "1px solid #e2e8f0", borderRadius: "9999px", padding: "0.15rem 0.5rem", fontWeight: 600 }}>{f.port}</span></span></td>
                  <td style={tdStyle}><span style={cellIconStyle}><Network size={13} /> {f.network.toUpperCase()}</span></td>
                  <td style={tdStyle}>
                    <button
                      onClick={() => handleDeleteForwarder(f.id)}
                      style={{
                        display: "flex",
                        alignItems: "center",
                        padding: "0.25rem",
                        background: "transparent",
                        border: "none",
                        color: "#dc2626",
                        cursor: "pointer",
                      }}>
                      <Trash2 size={14} />
                    </button>
                  </td>
                </tr>
              ))}
              {forwarders.length === 0 && (
                <tr>
                  <td
                    colSpan={4}
                    style={{
                      ...tdStyle,
                      color: "#64748b",
                      textAlign: "center",
                    }}>
                    No forwarders configured
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div style={{ display: "flex", justifyContent: "flex-end", marginTop: "0.75rem", paddingTop: "0.75rem" }}>
          <button
            onClick={() => setShowFwForm(true)}
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.3rem",
              padding: "0.6rem 1.2rem",
              background: colors.primary,
              color: "#fff",
              border: "none",
              borderRadius: "6px",
              cursor: "pointer",
              fontWeight: 600,
              fontSize: "0.85rem",
              fontFamily: "inherit",
            }}>
            <Plus size={14} />
            New Forwarder
          </button>
        </div>
      </div>

      {showFwForm && (
        <div
          onClick={() => setShowFwForm(false)}
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0,0,0,0.4)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
          }}>
          <div
            onClick={(e) => e.stopPropagation()}
            style={{
              background: "#fff",
              borderRadius: "8px",
              padding: "1.5rem",
              width: "100%",
              maxWidth: "400px",
              boxShadow: "0 4px 24px rgba(0,0,0,0.15)",
            }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: "1.25rem",
                paddingBottom: "0.75rem",
                borderBottom: `1px solid ${colors.border}`,
              }}>
              <h2
                style={{
                  margin: 0,
                  fontSize: "1rem",
                  fontWeight: 600,
                  display: "flex",
                  alignItems: "center",
                  gap: "0.5rem",
                }}>
                <Plus size={18} />
                New Forwarder
              </h2>
              <button
                type="button"
                onClick={() => setShowFwForm(false)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  padding: "0.3rem",
                  background: "transparent",
                  border: "none",
                  cursor: "pointer",
                  color: "#64748b",
                }}>
                <X size={18} />
              </button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
              <div>
                <label style={labelStyle}>Address</label>
                <input
                  type="text"
                  placeholder="8.8.8.8"
                  value={fwAddr}
                  onChange={(e) => setFwAddr(e.target.value)}
                  autoFocus
                  style={inputStyle}
                />
              </div>
              <div style={{ display: "flex", gap: "0.75rem" }}>
                <div style={{ flex: 1 }}>
                  <label style={labelStyle}>Port</label>
                  <input
                    type="number"
                    min={1}
                    max={65535}
                    placeholder="53"
                    value={fwPort || ""}
                    onChange={(e) => setFwPort(Number(e.target.value))}
                    style={inputStyle}
                  />
                </div>
                <div style={{ flex: 1 }}>
                  <label style={labelStyle}>Protocol</label>
                  <select
                    value={fwNetwork}
                    onChange={(e) => setFwNetwork(e.target.value)}
                    style={inputStyle}>
                    <option value="udp">UDP</option>
                    <option value="tcp">TCP</option>
                  </select>
                </div>
              </div>
              <button
                onClick={() => {
                  handleAddForwarder();
                  setShowFwForm(false);
                }}
                disabled={!fwAddr || !fwPort}
                style={{
                  alignSelf: "flex-end",
                  display: "flex",
                  alignItems: "center",
                  gap: "0.3rem",
                  padding: "0.6rem 1.2rem",
                  background: fwAddr && fwPort ? colors.primary : colors.border,
                  color: fwAddr && fwPort ? "#fff" : "#9ca3af",
                  border: "none",
                  borderRadius: "6px",
                  cursor: fwAddr && fwPort ? "pointer" : "not-allowed",
                  fontWeight: 600,
                  fontSize: "0.85rem",
                  fontFamily: "inherit",
                }}>
                Add
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

const cardStyle: React.CSSProperties = {
  width: "100%",
  background: "#fff",
  borderRadius: "8px",
  border: `1px solid ${colors.border}`,
  padding: "1.5rem",
  boxSizing: "border-box",
};

const cardTitleStyle: React.CSSProperties = {
  margin: "0 0 1.25rem",
  paddingBottom: "0.75rem",
  borderBottom: `1px solid ${colors.border}`,
  fontSize: "1rem",
  fontWeight: 600,
  display: "flex",
  alignItems: "center",
  gap: "0.5rem",
};

const labelStyle: React.CSSProperties = {
  display: "block",
  fontSize: "0.75rem",
  fontWeight: 600,
  textTransform: "uppercase",
  color: "#64748b",
  marginBottom: "0.3rem",
};

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: "0.6rem 0.8rem",
  background: "#f9fafb",
  border: "1px solid #e5e7eb",
  borderRadius: "6px",
  color: "#111",
  fontSize: "0.9rem",
  fontFamily: "inherit",
  boxSizing: "border-box",
};

const thStyle: React.CSSProperties = {
  padding: "0.5rem 0.75rem",
  fontWeight: 600,
  fontSize: "0.75rem",
  textTransform: "uppercase",
  color: "#64748b",
};

const thInnerStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.3rem",
};

const cellIconStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.4rem",
  color: "#64748b",
};

const tdStyle: React.CSSProperties = {
  padding: "0.5rem 0.75rem",
};
