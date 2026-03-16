import { useEffect, useState } from "react";
import {
  Users as UsersIcon,
  UserPlus,
  User,
  Calendar,
  Trash2,
  X,
  Plus,
} from "lucide-react";
import ConfirmModal from "./ConfirmModal";
import toast from "react-hot-toast";
import Avatar from "./Avatar";
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
    <div
      style={{
        padding: "2rem",
        display: "flex",
        flexDirection: "column",
        gap: "1.5rem",
      }}>
      {showForm && (
        <div
          onClick={() => {
            setShowForm(false);
            setError("");
          }}
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
                onClick={() => {
                  setShowForm(false);
                  setError("");
                }}
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
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "0.75rem",
              }}>
              <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  setError("");
                }}
                required
                autoFocus
                style={{
                  ...inputStyle,
                  width: "100%",
                  boxSizing: "border-box",
                }}
              />
              <input
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  setError("");
                }}
                required
                style={{
                  ...inputStyle,
                  width: "100%",
                  boxSizing: "border-box",
                }}
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
        <ConfirmModal
          title="Delete User"
          message={
            <>
              Are you sure you want to delete{" "}
              <strong>{deleteTarget.username}</strong>?
            </>
          }
          onConfirm={confirmDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}

      <div style={cardStyle}>
        <h2 style={cardTitleStyle}>
          <UsersIcon size={18} />
          Users
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
                    <User size={12} /> Username
                  </span>
                </th>
                <th style={thStyle}>
                  <span style={thInnerStyle}>
                    <Calendar size={12} /> Created
                  </span>
                </th>
                <th style={{ ...thStyle, width: "1%" }}></th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr
                  key={user.id}
                  style={{ borderBottom: `1px solid ${colors.border}` }}>
                  <td
                    style={{
                      ...tdStyle,
                      display: "flex",
                      alignItems: "center",
                      gap: "0.5rem",
                    }}>
                    <Avatar name={user.username} />
                    {user.username}
                  </td>
                  <td style={tdStyle}>
                    <span style={cellIconStyle}>
                      <Calendar size={13} />
                      {new Date(user.created_at).toLocaleDateString()}
                    </span>
                  </td>
                  <td style={tdStyle}>
                    <button
                      onClick={() => setDeleteTarget(user)}
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
              {users.length === 0 && (
                <tr>
                  <td
                    colSpan={3}
                    style={{
                      ...tdStyle,
                      color: "#64748b",
                      textAlign: "center",
                    }}>
                    No users
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            marginTop: "0.75rem",
            paddingTop: "0.75rem",
          }}>
          <button
            onClick={() => {
              setShowForm(true);
              setError("");
            }}
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
            Add User
          </button>
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
  padding: "0.6rem 0.8rem",
  background: "#f9fafb",
  border: `1px solid #e5e7eb`,
  borderRadius: "8px",
  color: "#111",
  fontSize: "0.9rem",
  fontFamily: "inherit",
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
