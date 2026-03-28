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
  BarChart3,
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
    time_series: { time: string; total: number; blocked: number; custom: number }[];
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

      {/* Queries over last 24 hours */}
      {data.stats?.time_series && <QueryChart series={data.stats.time_series} />}
    </div>
  );
}

function niceMax(val: number): number {
  if (val <= 0) return 10;
  const mag = Math.pow(10, Math.floor(Math.log10(val)));
  const norm = val / mag;
  if (norm <= 1) return mag;
  if (norm <= 2) return 2 * mag;
  if (norm <= 5) return 5 * mag;
  return 10 * mag;
}

function QueryChart({
  series,
}: {
  series: { time: string; total: number; blocked: number; custom: number }[];
}) {
  const rawMax = Math.max(...series.map((p) => p.total), 1);
  const maxVal = niceMax(rawMax);
  const gridLines = 4;
  const yLabels = Array.from({ length: gridLines + 1 }, (_, i) =>
    Math.round((maxVal / gridLines) * (gridLines - i))
  );

  const labelEvery = 12;
  const chartHeight = 160;
  const yLabelWidth = 40;

  return (
    <div
      style={{
        background: "#fff",
        borderRadius: "8px",
        border: `1px solid ${colors.border}`,
        padding: "1.25rem",
      }}>
      <div
        style={{
          fontSize: "0.7rem",
          fontWeight: 600,
          color: colors.navText,
          textTransform: "uppercase",
          letterSpacing: "0.03em",
          marginBottom: "1rem",
          display: "flex",
          alignItems: "center",
          gap: "0.4rem",
        }}>
        <BarChart3 size={14} />
        Queries over last 24 hours
        <div style={{ flex: 1 }} />
        <span
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: "0.25rem",
            textTransform: "none",
            fontWeight: 500,
          }}>
          <span
            style={{
              width: 8,
              height: 8,
              borderRadius: 2,
              background: "#16a34a",
              display: "inline-block",
            }}
          />
          Allowed
        </span>
        <span
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: "0.25rem",
            textTransform: "none",
            fontWeight: 500,
            marginLeft: "0.5rem",
          }}>
          <span
            style={{
              width: 8,
              height: 8,
              borderRadius: 2,
              background: "#dc2626",
              display: "inline-block",
            }}
          />
          Blocked
        </span>
        <span
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: "0.25rem",
            textTransform: "none",
            fontWeight: 500,
            marginLeft: "0.5rem",
          }}>
          <span
            style={{
              width: 8,
              height: 8,
              borderRadius: 2,
              background: "#3b82f6",
              display: "inline-block",
            }}
          />
          Custom
        </span>
      </div>

      <div style={{ display: "flex" }}>
        {/* Y-axis labels */}
        <div
          style={{
            width: yLabelWidth,
            height: chartHeight,
            display: "flex",
            flexDirection: "column",
            justifyContent: "space-between",
            flexShrink: 0,
          }}>
          {yLabels.map((v, i) => (
            <span
              key={i}
              style={{
                fontSize: "0.6rem",
                color: colors.navText,
                fontVariantNumeric: "tabular-nums",
                textAlign: "right",
                paddingRight: "0.5rem",
                lineHeight: 1,
              }}>
              {formatNumber(v)}
            </span>
          ))}
        </div>

        {/* Chart area with grid */}
        <div
          style={{
            flex: 1,
            height: chartHeight,
            position: "relative",
          }}>
          {/* Grid lines */}
          {yLabels.map((_, i) => (
            <div
              key={i}
              style={{
                position: "absolute",
                left: 0,
                right: 0,
                top: `${(i / gridLines) * 100}%`,
                borderTop: `1px solid ${i === gridLines ? colors.border : "#f0f0f0"}`,
              }}
            />
          ))}

          {/* Bars */}
          <div
            style={{
              display: "flex",
              alignItems: "flex-end",
              height: "100%",
              gap: 1,
              position: "relative",
              zIndex: 1,
            }}>
            {series.map((point, i) => {
              const blockedH = (point.blocked / maxVal) * 100;
              const customH = (point.custom / maxVal) * 100;
              const allowedH = ((point.total - point.blocked - point.custom) / maxVal) * 100;

              return (
                <div
                  key={i}
                  style={{
                    flex: "1 1 0",
                    minWidth: 0,
                    height: "100%",
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "flex-end",
                  }}>
                  <div
                    style={{
                      width: "100%",
                      height: `${Math.max(allowedH, 0)}%`,
                      background: "#16a34a",
                      borderRadius: "1px 1px 0 0",
                      minHeight: point.total - point.blocked - point.custom > 0 ? 1 : 0,
                    }}
                  />
                  <div
                    style={{
                      width: "100%",
                      height: `${customH}%`,
                      background: "#3b82f6",
                      minHeight: point.custom > 0 ? 1 : 0,
                    }}
                  />
                  <div
                    style={{
                      width: "100%",
                      height: `${blockedH}%`,
                      background: "#dc2626",
                      minHeight: point.blocked > 0 ? 1 : 0,
                    }}
                  />
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* X-axis labels */}
      <div
        style={{
          display: "flex",
          marginTop: "0.4rem",
          paddingTop: "0.3rem",
          paddingLeft: yLabelWidth,
          paddingRight: "1.5rem",
        }}>
        {series.map((point, i) => (
          <div
            key={i}
            style={{
              flex: "1 1 0",
              minWidth: 0,
              textAlign: "center",
              fontSize: "0.6rem",
              color: colors.navText,
              fontVariantNumeric: "tabular-nums",
            }}>
            {point.time.endsWith(":00") && parseInt(point.time) % 2 === 0
              ? point.time
              : ""}
          </div>
        ))}
      </div>
    </div>
  );
}
