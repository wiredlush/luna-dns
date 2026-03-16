import { useState } from "react";
import { colors } from "./theme";

export default function Login({ onLogin }: { onLogin: () => void }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await fetch("/api/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "same-origin",
        body: JSON.stringify({ username, password }),
      });

      if (!res.ok) {
        setError("Invalid username and/or password! Please try again.");
        return;
      }

      onLogin();
    } catch {
      setError("Network error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        alignItems: "center",
        height: "100%",
        background: "#fff",
      }}>
      <img
        src="/logo.svg"
        alt="Luna DNS"
        style={{ height: "200px", objectFit: "contain", marginBottom: "2rem" }}
      />
      <form
        onSubmit={handleSubmit}
        style={{
          display: "flex",
          flexDirection: "column",
          gap: "1rem",
          width: "320px",
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

        <div
          style={{
            padding: error ? "0.75rem 1rem" : "0",
            maxHeight: error ? "100px" : "0",
            opacity: error ? 1 : 0,
            overflow: "hidden",
            background: "rgba(239, 65, 54, 0.08)",
            border: error
              ? `1px solid ${colors.primary}`
              : "1px solid transparent",
            borderRadius: "8px",
            color: colors.primary,
            fontSize: "0.875rem",
            fontWeight: 500,
            transition: "all 0.3s ease",
          }}>
          {error}
        </div>

        <button
          type="submit"
          disabled={loading}
          style={{
            padding: "0.6rem 1.2rem",
            background: colors.primary,
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            cursor: loading ? "wait" : "pointer",
            fontWeight: 600,
          }}>
          {loading ? "Signing in..." : "Sign in"}
        </button>
      </form>
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  padding: "0.75rem 1rem",
  background: "#f9fafb",
  border: "1px solid #e5e7eb",
  borderRadius: "8px",
  color: "#111",
  fontSize: "0.95rem",
  transition: "border-color 0.2s, background 0.2s",
};
