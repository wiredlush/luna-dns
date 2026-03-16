import { useEffect, useState } from "react";
import { NavLink, Outlet } from "react-router-dom";
import {
  LayoutDashboard,
  Server,
  ShieldBan,
  ShieldCheck,
  Users,
  LogOut,
} from "lucide-react";
import { colors } from "./theme";

const navItems = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard },
  { to: "/dns", label: "DNS", icon: Server },
  { to: "/blocklist", label: "Blocklist", icon: ShieldBan },
  { to: "/security", label: "Security", icon: ShieldCheck },
  { to: "/users", label: "Users", icon: Users },
];

interface StatusData {
  dns: { running: boolean };
}

export default function Layout({ onLogout }: { onLogout: () => void }) {
  const [status, setStatus] = useState<StatusData | null>(null);

  useEffect(() => {
    let active = true;
    const poll = () => {
      fetch("/api/status", { credentials: "same-origin" })
        .then((r) => r.json())
        .then((data) => {
          if (active) setStatus(data);
        })
        .catch(() => {
          if (active) setStatus(null);
        });
    };
    poll();
    const id = setInterval(poll, 10000);
    return () => {
      active = false;
      clearInterval(id);
    };
  }, []);

  async function handleLogout() {
    await fetch("/api/logout", { method: "POST", credentials: "same-origin" });
    onLogout();
  }

  const dnsOnline = status?.dns?.running ?? false;

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <header
        style={{
          display: "flex",
          alignItems: "center",
          height: "60px",
          background: colors.navbar,
          borderBottom: `1px solid ${colors.border}`,
          flexShrink: 0,
        }}>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            maxWidth: "1200px",
            margin: "0 auto",
            padding: "0 1.5rem",
            width: "100%",
            height: "100%",
          }}>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.75rem",
              marginRight: "1rem",
            }}>
            <img
              src="/logo.svg"
              alt="Luna DNS"
              style={{ height: "45px", objectFit: "contain" }}
            />
          </div>

          <nav style={{ display: "flex", gap: "4px", flex: 1 }}>
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === "/"}
                style={({ isActive }) => ({
                  display: "flex",
                  alignItems: "center",
                  gap: "0.4rem",
                  padding: "0.4rem 0.9rem",
                  borderRadius: "6px",
                  textDecoration: "none",
                  fontSize: "0.85rem",
                  fontWeight: 500,
                  color: isActive ? colors.primary : colors.navText,
                  background: isActive ? "rgba(239,65,54,0.08)" : "transparent",
                })}>
                <item.icon size={16} />
                {item.label}
              </NavLink>
            ))}
          </nav>

          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.4rem",
              marginRight: "1rem",
              fontSize: "0.7rem",
              fontWeight: 600,
              color: dnsOnline ? "#16a34a" : "#dc2626",
              background: "transparent",
              border: `1px solid ${colors.border}`,
              borderRadius: "6px",
              padding: "0.4rem 0.6rem",
            }}>
            <span
              style={{
                display: "inline-block",
                width: 8,
                height: 8,
                borderRadius: "50%",
                background: dnsOnline ? "#16a34a" : "#dc2626",
                animation: dnsOnline
                  ? "pulse-dot 1.5s ease-in-out infinite"
                  : "none",
              }}
            />
            DNS {dnsOnline ? "Online" : "Offline"}
          </div>

          <button
            onClick={handleLogout}
            title="Logout"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              padding: "0.5rem",
              background: "transparent",
              border: `1px solid ${colors.border}`,
              borderRadius: "6px",
              color: colors.navText,
              cursor: "pointer",
            }}>
            <LogOut size={16} />
          </button>
        </div>
      </header>

      <main style={{ flex: 1, overflow: "auto", background: colors.bg }}>
        <div
          style={{ maxWidth: "1200px", margin: "0 auto", padding: "0 1.5rem" }}>
          <Outlet />
        </div>
      </main>
    </div>
  );
}
