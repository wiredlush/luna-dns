import { useEffect, useState } from "react";
import { UserPlus, Trash2, X } from "lucide-react";
import toast from "react-hot-toast";
import { colors } from "./theme";

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

interface User {
  id: number;
  username: string;
  created_at: string;
}

export default function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null);

  useEffect(() => {
    fetchUsers();
  }, []);

  async function fetchUsers() {
    const res = await fetch("/api/users", { credentials: "same-origin" });
    if (res.ok) setUsers(await res.json());
  }

  async function handleCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");

    const res = await fetch("/api/users", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "same-origin",
      body: JSON.stringify({ username, password }),
    });

    if (!res.ok) {
      const data = await res.json();
      setError(capitalize(data.error || "Failed to create user"));
      return;
    }

    toast.success("User created successfully");
    setUsername("");
    setPassword("");
    setShowForm(false);
    fetchUsers();
  }

  async function confirmDelete() {
    if (!deleteTarget) return;

    const res = await fetch(`/api/users/${deleteTarget.id}`, {
      method: "DELETE",
      credentials: "same-origin",
    });

    if (!res.ok) {
      const data = await res.json();
      toast.error(capitalize(data.error || "Failed to delete user"));
      setDeleteTarget(null);
      return;
    }

    toast.success("User deleted successfully");
    setDeleteTarget(null);
    fetchUsers();
  }

  return (
    <div style={{ padding: "2rem" }}>
      {showForm && (
        <div
          onClick={() => { setShowForm(false); setError(""); }}
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
                <UserPlus size={18} />
                Add User
              </h2>
              <button
                type="button"
                onClick={() => { setShowForm(false); setError(""); }}
                style={{
                  display: "flex",
                  alignItems: "center",
                  padding: "0.3rem",
                  background: "transparent",
                  border: "none",
                  cursor: "pointer",
                  color: colors.navText,
                }}>
                <X size={18} />
              </button>
            </div>
            <form
              onSubmit={handleCreate}
              style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
              <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => { setUsername(e.target.value); setError(""); }}
                required
                autoFocus
                style={{ ...inputStyle, width: "100%", boxSizing: "border-box" }}
              />
              <input
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => { setPassword(e.target.value); setError(""); }}
                required
                style={{ ...inputStyle, width: "100%", boxSizing: "border-box" }}
              />
              {error && (
                <span style={{ color: "#dc2626", fontSize: "0.85rem" }}>
                  {error}
                </span>
              )}
              <button
                type="submit"
                style={{
                  alignSelf: "flex-end",
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
                Create
              </button>
            </form>
          </div>
        </div>
      )}

      {deleteTarget && (
        <div
          onClick={() => setDeleteTarget(null)}
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
              maxWidth: "360px",
              boxShadow: "0 4px 24px rgba(0,0,0,0.15)",
            }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: "1rem",
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
                <Trash2 size={18} />
                Delete User
              </h2>
              <button
                type="button"
                onClick={() => setDeleteTarget(null)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  padding: "0.3rem",
                  background: "transparent",
                  border: "none",
                  cursor: "pointer",
                  color: colors.navText,
                }}>
                <X size={18} />
              </button>
            </div>
            <p style={{ margin: "0 0 1.25rem", fontSize: "0.9rem" }}>
              Are you sure you want to delete <strong>{deleteTarget.username}</strong>?
            </p>
            <div style={{ display: "flex", justifyContent: "flex-end", gap: "0.5rem" }}>
              <button
                onClick={() => setDeleteTarget(null)}
                style={{
                  padding: "0.5rem 1rem",
                  background: "transparent",
                  border: `1px solid ${colors.border}`,
                  borderRadius: "6px",
                  cursor: "pointer",
                  fontSize: "0.85rem",
                  fontFamily: "inherit",
                }}>
                Cancel
              </button>
              <button
                onClick={confirmDelete}
                style={{
                  padding: "0.5rem 1rem",
                  background: colors.primary,
                  color: "#fff",
                  border: "none",
                  borderRadius: "6px",
                  cursor: "pointer",
                  fontWeight: 600,
                  fontSize: "0.85rem",
                  fontFamily: "inherit",
                }}>
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      <div
        style={{
          background: "#fff",
          borderRadius: "8px",
          border: `1px solid ${colors.border}`,
          overflow: "auto",
        }}>
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            fontSize: "0.9rem",
          }}>
          <thead>
            <tr
              style={{
                borderBottom: `1px solid ${colors.border}`,
                textAlign: "left",
              }}>
              <th style={thStyle}>Username</th>
              <th style={thStyle}>Created</th>
              <th style={{ ...thStyle, width: "1%", whiteSpace: "nowrap" }}>
                <button
                  onClick={() => {
                    setShowForm(!showForm);
                    setError("");
                  }}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.3rem",
                    padding: "0.3rem 0.6rem",
                    background: colors.primary,
                    color: "#fff",
                    border: `1px solid ${colors.primary}`,
                    borderRadius: "4px",
                    cursor: "pointer",
                    fontSize: "0.8rem",
                    fontWeight: 600,
                    fontFamily: "inherit",
                  }}>
                  <UserPlus size={14} />
                </button>
              </th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <tr
                key={user.id}
                style={{ borderBottom: `1px solid ${colors.border}` }}>
                <td style={tdStyle}>{user.username}</td>
                <td style={tdStyle}>
                  {new Date(user.created_at).toLocaleDateString()}
                </td>
                <td style={tdStyle}>
                  <button
                    onClick={() => setDeleteTarget(user)}
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
                    <Trash2 size={14} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

    </div>
  );
}

const inputStyle: React.CSSProperties = {
  padding: "0.6rem 0.8rem",
  background: "#f9fafb",
  border: `1px solid #e5e7eb`,
  borderRadius: "8px",
  color: "#111",
  fontSize: "0.9rem",
  fontFamily: "inherit",
};

const thStyle: React.CSSProperties = {
  padding: "0.75rem 1rem",
  fontWeight: 600,
  fontSize: "0.8rem",
  textTransform: "uppercase",
  color: "#64748b",
};

const tdStyle: React.CSSProperties = {
  padding: "0.75rem 1rem",
};
