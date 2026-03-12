import { useEffect, useState } from "react";
import { colors } from "./theme";

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
      setError(data.error || "Failed to create user");
      return;
    }

    setUsername("");
    setPassword("");
    setShowForm(false);
    fetchUsers();
  }

  async function handleDelete(id: number) {
    if (!confirm("Are you sure you want to delete this user?")) return;

    const res = await fetch(`/api/users/${id}`, {
      method: "DELETE",
      credentials: "same-origin",
    });

    if (!res.ok) {
      const data = await res.json();
      alert(data.error || "Failed to delete user");
      return;
    }

    fetchUsers();
  }

  return (
    <div style={{ padding: "2rem" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "1.5rem",
        }}>
        <h1 style={{ margin: 0, fontSize: "1.5rem", fontWeight: 600 }}>
          Users
        </h1>
        <button
          onClick={() => {
            setShowForm(!showForm);
            setError("");
          }}
          style={{
            padding: "0.5rem 1rem",
            background: colors.primary,
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
            fontSize: "0.85rem",
            fontWeight: 600,
            fontFamily: "inherit",
          }}>
          {showForm ? "Cancel" : "Add User"}
        </button>
      </div>

      {showForm && (
        <form
          onSubmit={handleCreate}
          style={{
            display: "flex",
            gap: "0.75rem",
            alignItems: "flex-start",
            marginBottom: "1.5rem",
            padding: "1rem",
            background: "#fff",
            borderRadius: "8px",
            border: `1px solid ${colors.border}`,
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
            style={inputStyle}
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
            style={inputStyle}
          />
          <button
            type="submit"
            style={{
              padding: "0.6rem 1.2rem",
              background: colors.primary,
              color: "#fff",
              border: "none",
              borderRadius: "8px",
              cursor: "pointer",
              fontWeight: 600,
              fontFamily: "inherit",
              whiteSpace: "nowrap",
            }}>
            Create
          </button>
          {error && (
            <span
              style={{
                color: colors.primary,
                fontSize: "0.85rem",
                alignSelf: "center",
              }}>
              {error}
            </span>
          )}
        </form>
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
              <th style={{ ...thStyle, width: "80px" }}></th>
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
                    onClick={() => handleDelete(user.id)}
                    style={{
                      padding: "0.3rem 0.6rem",
                      background: "transparent",
                      color: colors.primary,
                      border: `1px solid ${colors.primary}`,
                      borderRadius: "4px",
                      cursor: "pointer",
                      fontSize: "0.8rem",
                      fontFamily: "inherit",
                    }}>
                    Delete
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
