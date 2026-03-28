import { useEffect, useState } from "react";
import {
  Activity,
  ShieldBan,
  FileText,
  Users,
  Globe,
  Server,
  Cpu,
  MemoryStick,
  HardDrive,
  Clock,
  Database,
  Target,
} from "lucide-react";
import { colors } from "./theme";

interface StatusData {
  dns: { running: boolean; uptime_sec: number; cache_size: number };
  stats: {
    total_queries: number;
    blocked_queries: number;
    custom_queries: number;
    cache_hits: number;
    cache_misses: number;
    cache_hit_rate: number;
    unique_clients: number;
    unique_domains: number;
  };
  system: {
    cpu: number;
    ram: number;
    disk: number;
  };
}

function formatUptime(sec: number): string {
  if (sec < 60) return `${sec}s`;
  const m = Math.floor(sec / 60) % 60;
  const h = Math.floor(sec / 3600) % 24;
  const d = Math.floor(sec / 86400);
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

function formatNumber(n: number): string {
  if (n >= 1_000_000_000) return (n / 1_000_000_000).toFixed(2) + "B";
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(2) + "M";
  if (n >= 10_000) return (n / 1_000).toFixed(1) + "K";
  return n.toLocaleString();
}

const cards = [
  { key: "total_queries", label: "Total Queries", icon: Activity, accent: colors.primary },
  { key: "custom_queries", label: "Custom Records", icon: FileText, accent: "#3b82f6" },
  { key: "blocked_queries", label: "Blocked", icon: ShieldBan, accent: "#dc2626" },
  { key: "unique_clients", label: "Clients", icon: Users, accent: "#8b5cf6" },
  { key: "unique_domains", label: "Domains", icon: Globe, accent: "#0891b2" },
] as const;

export default function Dashboard() {
  const [data, setData] = useState<StatusData | null>(null);

  useEffect(() => {
    let active = true;
    const poll = () => {
      fetch("/api/status", { credentials: "same-origin" })
        .then((r) => r.json())
        .then((d) => {
          if (active) setData(d);
        })
        .catch(() => {});
    };
    poll();
    const id = setInterval(poll, 2000);
    return () => {
      active = false;
      clearInterval(id);
    };
  }, []);

  if (!data) return null;

  const dnsOnline = data.dns?.running ?? false;
  const sys = data.system;

  const sysItems = [
    { label: "CPU", value: `${sys?.cpu ?? 0}%`, icon: Cpu },
    { label: "RAM", value: `${sys?.ram ?? 0}%`, icon: MemoryStick },
    { label: "Disk", value: `${sys?.disk ?? 0}%`, icon: HardDrive },
  ];

  return (
    <div
      style={{
        padding: "2rem",
        display: "flex",
        flexDirection: "column",
        gap: "1rem",
      }}>
      {/* System status bar */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "1.5rem",
          background: "#fff",
          borderRadius: "8px",
          border: `1px solid ${colors.border}`,
          padding: "0.65rem 1.25rem",
        }}>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "0.4rem",
            fontSize: "0.75rem",
            fontWeight: 600,
            color: dnsOnline ? "#16a34a" : "#dc2626",
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

        {dnsOnline && (
          <>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.35rem",
                fontSize: "0.75rem",
                fontWeight: 500,
                color: colors.navText,
              }}>
              <Clock size={13} />
              <span style={{ fontWeight: 600 }}>Uptime</span>
              <span style={{ fontVariantNumeric: "tabular-nums" }}>
                {formatUptime(data.dns?.uptime_sec ?? 0)}
              </span>
            </div>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.35rem",
                fontSize: "0.75rem",
                fontWeight: 500,
                color: colors.navText,
              }}>
              <Database size={13} />
              <span style={{ fontWeight: 600 }}>Cache</span>
              <span style={{ fontVariantNumeric: "tabular-nums" }}>
                {formatNumber(data.dns?.cache_size ?? 0)}
              </span>
            </div>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.35rem",
                fontSize: "0.75rem",
                fontWeight: 500,
                color: colors.navText,
              }}>
              <Target size={13} />
              <span style={{ fontWeight: 600 }}>Hit Rate</span>
              <span style={{ fontVariantNumeric: "tabular-nums" }}>
                {data.stats?.cache_hit_rate ?? 0}%
              </span>
            </div>
          </>
        )}

        <div
          style={{
            width: 1,
            height: 16,
            background: colors.border,
          }}
        />

        {sysItems.map((item) => (
          <div
            key={item.label}
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.35rem",
              fontSize: "0.75rem",
              fontWeight: 500,
              color: colors.navText,
            }}>
            <item.icon size={13} />
            <span style={{ fontWeight: 600 }}>{item.label}</span>
            <span style={{ fontVariantNumeric: "tabular-nums" }}>
              {item.value}
            </span>
          </div>
        ))}
      </div>

      {/* Stats cards */}
      <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap" }}>
        {cards.map((card) => {
          const value = data.stats?.[card.key] ?? 0;
          return (
            <div
              key={card.key}
              style={{
                flex: "1 1 0",
                minWidth: 0,
                background: "#fff",
                borderRadius: "8px",
                border: `1px solid ${colors.border}`,
                padding: "1rem 1.25rem",
                display: "flex",
                alignItems: "center",
                gap: "1rem",
              }}>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  width: 36,
                  height: 36,
                  borderRadius: "8px",
                  background: card.accent,
                  color: "#fff",
                  flexShrink: 0,
                }}>
                <card.icon size={18} />
              </div>
              <div>
                <div
                  style={{
                    fontSize: "0.7rem",
                    fontWeight: 600,
                    color: colors.navText,
                    textTransform: "uppercase",
                    letterSpacing: "0.03em",
                    marginBottom: "0.15rem",
                    whiteSpace: "nowrap",
                  }}>
                  {card.label}
                </div>
                <div
                  style={{
                    fontSize: "1.5rem",
                    fontWeight: 700,
                    color: colors.text,
                    fontVariantNumeric: "tabular-nums",
                    lineHeight: 1,
                  }}>
                  {formatNumber(value)}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
