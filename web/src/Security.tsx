import { useEffect, useState } from "react";
import { Lock, Monitor, ScrollText, LogOut } from "lucide-react";
import { colors } from "./theme";

type Strength = "weak" | "fair" | "good" | "strong";

function getStrength(pw: string): Strength {
  let score = 0;
  if (pw.length >= 8) score++;
  if (pw.length >= 12) score++;
  if (/[a-z]/.test(pw) && /[A-Z]/.test(pw)) score++;
  if (/\d/.test(pw)) score++;
  if (/[^a-zA-Z0-9]/.test(pw)) score++;
  if (score <= 1) return "weak";
  if (score <= 2) return "fair";
  if (score <= 3) return "good";
  return "strong";
}

const strengthColor: Record<Strength, string> = {
  weak: "#dc2626",
  fair: "#f59e0b",
  good: "#3b82f6",
  strong: "#16a34a",
};

const strengthWidth: Record<Strength, string> = {
  weak: "25%",
  fair: "50%",
  good: "75%",
  strong: "100%",
};

interface SessionInfo {
  username: string;
  ip: string;
  created_at: string;
  expires: string;
  current: boolean;
}

interface AuditEntry {
  id: number;
  timestamp: string;
  username: string;
  action: string;
  detail: string;
  ip: string;
}

const actionColors: Record<string, string> = {
  login: "#16a34a",
  login_failed: "#dc2626",
  logout: "#64748b",
  logout_all: "#64748b",
  password_change: "#f59e0b",
  user_create: "#3b82f6",
  user_delete: "#dc2626",
};

export default function Security() {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  const [sessions, setSessions] = useState<SessionInfo[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditEntry[]>([]);

  useEffect(() => {
    fetchSessions();
    fetchAuditLogs();
  }, []);

  function fetchSessions() {
    fetch("/api/sessions", { credentials: "same-origin" })
      .then((r) => r.json())
      .then((data) => setSessions(data || []))
      .catch(() => {});
  }

  function fetchAuditLogs() {
    fetch("/api/audit-logs?limit=50", { credentials: "same-origin" })
      .then((r) => r.json())
      .then((data) => setAuditLogs(data || []))
      .catch(() => {});
  }

  async function handleLogoutAll() {
    await fetch("/api/sessions/logout-all", {
      method: "POST",
      credentials: "same-origin",
    });
    fetchSessions();
    fetchAuditLogs();
  }

  const strength = newPassword ? getStrength(newPassword) : null;
  const mismatch = confirmPassword !== "" && newPassword !== confirmPassword;
  const tooShort = newPassword !== "" && newPassword.length < 8;
  const canSubmit =
    currentPassword !== "" &&
    newPassword.length >= 8 &&
    newPassword === confirmPassword &&
    !loading;

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setSuccess("");
    setLoading(true);

    try {
      const res = await fetch("/api/change-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "same-origin",
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword,
        }),
      });

      if (!res.ok) {
        const data = await res.json();
        setError(data.error || "Failed to change password");
        return;
      }

      setSuccess("Password changed successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      fetchAuditLogs();
    } catch {
      setError("Network error");
    } finally {
      setLoading(false);
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
      {/* Change Password */}
      <div style={cardStyle}>
        <h2 style={cardTitleStyle}>
          <Lock size={18} />
          Change Password
        </h2>
        <form
          onSubmit={handleSubmit}
          style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
          <input
            type="password"
            placeholder="Current password"
            value={currentPassword}
            onChange={(e) => {
              setCurrentPassword(e.target.value);
              setError("");
              setSuccess("");
            }}
            required
            style={inputStyle}
          />
          <div>
            <input
              type="password"
              placeholder="New password"
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value);
                setError("");
                setSuccess("");
              }}
              required
              style={inputStyle}
            />
            {tooShort && (
              <span style={{ fontSize: "0.75rem", color: "#dc2626" }}>
                Must be at least 8 characters
              </span>
            )}
            {strength && !tooShort && (
              <div style={{ marginTop: "0.4rem" }}>
                <div
                  style={{
                    height: 4,
                    borderRadius: 2,
                    background: colors.border,
                  }}>
                  <div
                    style={{
                      height: "100%",
                      width: strengthWidth[strength],
                      borderRadius: 2,
                      background: strengthColor[strength],
                      transition: "width 0.2s, background 0.2s",
                    }}
                  />
                </div>
                <span
                  style={{
                    fontSize: "0.7rem",
                    color: strengthColor[strength],
                    textTransform: "capitalize",
                  }}>
                  {strength}
                </span>
              </div>
            )}
          </div>
          <div>
            <input
              type="password"
              placeholder="Confirm new password"
              value={confirmPassword}
              onChange={(e) => {
                setConfirmPassword(e.target.value);
                setError("");
                setSuccess("");
              }}
              required
              style={{
                ...inputStyle,
                borderColor: mismatch ? "#dc2626" : "#e5e7eb",
              }}
            />
            {mismatch && (
              <span style={{ fontSize: "0.75rem", color: "#dc2626" }}>
                Passwords do not match
              </span>
            )}
          </div>
          {error && (
            <span style={{ fontSize: "0.85rem", color: "#dc2626" }}>
              {error}
            </span>
          )}
          {success && (
            <span style={{ fontSize: "0.85rem", color: "#16a34a" }}>
              {success}
            </span>
          )}
          <button
            type="submit"
            disabled={!canSubmit}
            style={{
              alignSelf: "flex-end",
              padding: "0.6rem 1.2rem",
              background: canSubmit ? colors.primary : colors.border,
              color: canSubmit ? "#fff" : "#9ca3af",
              border: "none",
              borderRadius: "6px",
              cursor: canSubmit ? "pointer" : "not-allowed",
              fontWeight: 600,
              fontSize: "0.85rem",
              fontFamily: "inherit",
            }}>
            {loading ? "Saving..." : "Change Password"}
          </button>
        </form>
      </div>

      {/* Active Sessions */}
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
            <Monitor size={18} />
            Active Sessions
          </h2>
          <button
            onClick={handleLogoutAll}
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.3rem",
              padding: "0.3rem 0.6rem",
              background: "transparent",
              color: colors.primary,
              border: `1px solid ${colors.primary}`,
              borderRadius: "4px",
              cursor: "pointer",
              fontSize: "0.8rem",
              fontFamily: "inherit",
            }}>
            <LogOut size={14} />
            Logout Others
          </button>
        </div>
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
                <th style={thStyle}>User</th>
                <th style={thStyle}>IP</th>
                <th style={thStyle}>Created</th>
                <th style={thStyle}>Expires</th>
                <th style={{ ...thStyle, width: "1%" }}></th>
              </tr>
            </thead>
            <tbody>
              {sessions.map((s, i) => (
                <tr
                  key={i}
                  style={{
                    borderBottom: `1px solid ${colors.border}`,
                    background: s.current ? "rgba(239,65,54,0.04)" : undefined,
                  }}>
                  <td style={tdStyle}>{s.username}</td>
                  <td style={tdStyle}>{s.ip}</td>
                  <td style={tdStyle}>
                    {new Date(s.created_at).toLocaleString()}
                  </td>
                  <td style={tdStyle}>
                    {new Date(s.expires).toLocaleString()}
                  </td>
                  <td style={tdStyle}>
                    {s.current && (
                      <span
                        style={{
                          fontSize: "0.7rem",
                          color: "#16a34a",
                          fontWeight: 600,
                          whiteSpace: "nowrap",
                        }}>
                        current
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Audit Log */}
      <div style={cardStyle}>
        <h2 style={cardTitleStyle}>
          <ScrollText size={18} />
          Audit Log
        </h2>
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
                <th style={thStyle}>Timestamp</th>
                <th style={thStyle}>User</th>
                <th style={thStyle}>Action</th>
                <th style={thStyle}>Detail</th>
                <th style={thStyle}>IP</th>
              </tr>
            </thead>
            <tbody>
              {auditLogs.map((log) => (
                <tr
                  key={log.id}
                  style={{ borderBottom: `1px solid ${colors.border}` }}>
                  <td style={{ ...tdStyle, whiteSpace: "nowrap" }}>
                    {new Date(log.timestamp).toLocaleString()}
                  </td>
                  <td style={tdStyle}>{log.username}</td>
                  <td style={tdStyle}>
                    <span
                      style={{
                        color: actionColors[log.action] || colors.text,
                        fontWeight: 500,
                      }}>
                      {log.action}
                    </span>
                  </td>
                  <td style={tdStyle}>{log.detail || "—"}</td>
                  <td style={tdStyle}>{log.ip}</td>
                </tr>
              ))}
              {auditLogs.length === 0 && (
                <tr>
                  <td
                    colSpan={5}
                    style={{ ...tdStyle, color: "#64748b", textAlign: "center" }}>
                    No entries yet
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
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

const tdStyle: React.CSSProperties = {
  padding: "0.5rem 0.75rem",
};
