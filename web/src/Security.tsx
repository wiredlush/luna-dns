import { useEffect, useState } from "react";
import {
  Lock,
  KeyRound,
  Monitor,
  ScrollText,
  LogOut,
  ChevronLeft,
  ChevronRight,
  User,
  Globe,
  Calendar,
  Clock,
  Activity,
  FileText,
} from "lucide-react";
import toast from "react-hot-toast";
import Avatar from "./Avatar";
import Badge from "./Badge";
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

const actionLabels: Record<string, string> = {
  login: "Login",
  login_failed: "Login Failed",
  logout: "Logout",
  logout_all: "Logout All",
  password_change: "Password Change",
  user_create: "User Create",
  user_delete: "User Delete",
};

export default function Security() {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);

  const [sessions, setSessions] = useState<SessionInfo[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditEntry[]>([]);
  const [auditPage, setAuditPage] = useState(0);
  const auditPerPage = 10;

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
    fetch("/api/audit-logs", { credentials: "same-origin" })
      .then((r) => r.json())
      .then((data) => {
        setAuditLogs(data || []);
        setAuditPage(0);
      })
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

  const totalPages = Math.ceil(auditLogs.length / auditPerPage);
  const pagedLogs = auditLogs.slice(
    auditPage * auditPerPage,
    (auditPage + 1) * auditPerPage,
  );

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
        toast.error(data.error || "Failed to change password");
        return;
      }

      toast.success("Password changed successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      fetchAuditLogs();
    } catch {
      toast.error("Network error");
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
            onChange={(e) => setCurrentPassword(e.target.value)}
            required
            style={inputStyle}
          />
          <div>
            <input
              type="password"
              placeholder="New password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
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
              onChange={(e) => setConfirmPassword(e.target.value)}
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
          <div
            style={{
              display: "flex",
              justifyContent: "flex-end",
              paddingTop: "0.75rem",
            }}>
            <button
              type="submit"
              disabled={!canSubmit}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.3rem",
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
              <KeyRound size={14} />
              {loading ? "Saving..." : "Change Password"}
            </button>
          </div>
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
              padding: "0.6rem 1.2rem",
              background: "transparent",
              color: colors.primary,
              border: `1px solid ${colors.primary}`,
              borderRadius: "6px",
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
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <User size={12} /> User
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <Globe size={12} /> IP
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <Calendar size={12} /> Created
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <Clock size={12} /> Expires
                  </span>
                </th>
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
                  <td
                    style={{
                      ...tdStyle,
                      display: "flex",
                      alignItems: "center",
                      gap: "0.5rem",
                    }}>
                    <Avatar name={s.username} />
                    {s.username}
                  </td>
                  <td style={tdStyle}>
                    <Badge mono>{s.ip}</Badge>
                  </td>
                  <td style={tdStyle}>
                    <span style={cellIconStyle}>
                      <Calendar size={13} />{" "}
                      {new Date(s.created_at).toLocaleString()}
                    </span>
                  </td>
                  <td style={tdStyle}>
                    <span style={cellIconStyle}>
                      <Clock size={13} /> {new Date(s.expires).toLocaleString()}
                    </span>
                  </td>
                  <td style={tdStyle}>
                    {s.current && <Badge color="#16a34a">Current</Badge>}
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
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <Calendar size={12} /> Timestamp
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <User size={12} /> User
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <Activity size={12} /> Action
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <FileText size={12} /> Detail
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <Globe size={12} /> IP
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              {pagedLogs.map((log) => (
                <tr
                  key={log.id}
                  style={{ borderBottom: `1px solid ${colors.border}` }}>
                  <td style={{ ...tdStyle, whiteSpace: "nowrap" }}>
                    <span style={cellIconStyle}>
                      <Calendar size={13} />{" "}
                      {new Date(log.timestamp).toLocaleString()}
                    </span>
                  </td>
                  <td
                    style={{
                      ...tdStyle,
                      display: "flex",
                      alignItems: "center",
                      gap: "0.5rem",
                    }}>
                    <Avatar name={log.username} />
                    {log.username}
                  </td>
                  <td style={tdStyle}>
                    <Badge color={actionColors[log.action] || colors.text}>
                      {actionLabels[log.action] || log.action}
                    </Badge>
                  </td>
                  <td style={tdStyle}>
                    <span style={cellIconStyle}>
                      <FileText size={13} /> {log.detail || "—"}
                    </span>
                  </td>
                  <td style={tdStyle}>
                    <Badge mono>{log.ip}</Badge>
                  </td>
                </tr>
              ))}
              {pagedLogs.length === 0 && (
                <tr>
                  <td
                    colSpan={5}
                    style={{
                      ...tdStyle,
                      color: "#64748b",
                      textAlign: "center",
                    }}>
                    No entries yet
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {totalPages > 1 && (
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginTop: "0.75rem",
              paddingTop: "0.75rem",
              borderTop: `1px solid ${colors.border}`,
              fontSize: "0.8rem",
              color: "#64748b",
            }}>
            <span>
              {auditPage * auditPerPage + 1}–
              {Math.min((auditPage + 1) * auditPerPage, auditLogs.length)} of{" "}
              {auditLogs.length}
            </span>
            <div style={{ display: "flex", gap: "0.25rem" }}>
              <button
                onClick={() => setAuditPage((p) => p - 1)}
                disabled={auditPage === 0}
                style={paginationBtnStyle(auditPage === 0)}>
                <ChevronLeft size={14} />
              </button>
              <button
                onClick={() => setAuditPage((p) => p + 1)}
                disabled={auditPage + 1 >= totalPages}
                style={paginationBtnStyle(auditPage + 1 >= totalPages)}>
                <ChevronRight size={14} />
              </button>
            </div>
          </div>
        )}
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

function paginationBtnStyle(disabled: boolean): React.CSSProperties {
  return {
    display: "flex",
    alignItems: "center",
    padding: "0.3rem",
    background: "transparent",
    border: `1px solid ${disabled ? colors.border : "#d1d5db"}`,
    borderRadius: "4px",
    color: disabled ? "#d1d5db" : "#374151",
    cursor: disabled ? "not-allowed" : "pointer",
  };
}
