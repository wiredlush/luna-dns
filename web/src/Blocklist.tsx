import { useEffect, useState, useRef } from "react";
import {
  ShieldBan,
  Plus,
  Upload,
  Trash2,
  X,
  Link,
  Search,
  ChevronLeft,
  ChevronRight,
  Loader2,
} from "lucide-react";
import toast from "react-hot-toast";
import Badge from "./Badge";
import ConfirmModal from "./ConfirmModal";
import { colors } from "./theme";

interface BlocklistEntry {
  id: number;
  domain: string;
}

const perPage = 15;

export default function Blocklist() {
  const [entries, setEntries] = useState<BlocklistEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [search, setSearch] = useState("");
  const [loaded, setLoaded] = useState(false);
  const [domain, setDomain] = useState("");
  const [showAddForm, setShowAddForm] = useState(false);
  const [showUploadForm, setShowUploadForm] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [showClearConfirm, setShowClearConfirm] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  function fetchEntries(p: number, q: string) {
    const params = new URLSearchParams({
      limit: String(perPage),
      offset: String(p * perPage),
    });
    if (q) params.set("search", q);

    fetch(`/api/blocklist?${params}`, { credentials: "same-origin" })
      .then((r) => r.json())
      .then((data) => {
        setEntries(data.entries || []);
        setTotal(data.total || 0);
        setLoaded(true);
      })
      .catch(() => {});
  }

  useEffect(() => {
    fetchEntries(0, "");
  }, []);

  function handleSearch(value: string) {
    setSearch(value);
    setPage(0);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => fetchEntries(0, value), 300);
  }

  function goToPage(p: number) {
    setPage(p);
    fetchEntries(p, search);
  }

  async function handleAdd() {
    if (!domain) return;
    try {
      const res = await fetch("/api/blocklist", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "same-origin",
        body: JSON.stringify({ domain: domain.trim().toLowerCase() }),
      });
      const data = await res.json();
      if (!res.ok) {
        toast.error(data.error || "Failed to add entry");
        return;
      }
      setDomain("");
      toast.success("Entry added");
      fetchEntries(page, search);
    } catch {
      toast.error("Network error");
    }
  }

  async function handleDelete(id: number) {
    try {
      const res = await fetch(`/api/blocklist/${id}`, {
        method: "DELETE",
        credentials: "same-origin",
      });
      if (!res.ok) {
        const data = await res.json();
        toast.error(data.error || "Failed to delete entry");
        return;
      }
      toast.success("Blocklist entry removed");
      fetchEntries(page, search);
    } catch {
      toast.error("Network error");
    }
  }

  async function handleClearAll() {
    try {
      const res = await fetch("/api/blocklist", {
        method: "DELETE",
        credentials: "same-origin",
      });
      if (!res.ok) {
        const data = await res.json();
        toast.error(data.error || "Failed to clear blocklist");
        return;
      }
      toast.success("Blocklist cleared");
      setShowClearConfirm(false);
      setSearch("");
      setPage(0);
      fetchEntries(0, "");
    } catch {
      toast.error("Network error");
    }
  }

  async function handleUpload() {
    const file = fileRef.current?.files?.[0];
    if (!file) return;

    const form = new FormData();
    form.append("file", file);

    setUploading(true);
    try {
      const res = await fetch("/api/blocklist/upload", {
        method: "POST",
        credentials: "same-origin",
        body: form,
      });
      const data = await res.json();
      if (!res.ok) {
        toast.error(data.error || "Failed to upload file");
        return;
      }
      toast.success(`Imported ${data.imported} entries`);
      setShowUploadForm(false);
      if (fileRef.current) fileRef.current.value = "";
      setSearch("");
      setPage(0);
      fetchEntries(0, "");
    } catch {
      toast.error("Unable to upload file");
    } finally {
      setUploading(false);
    }
  }

  const totalPages = Math.ceil(total / perPage);

  if (!loaded) {
    return <div style={{ padding: "2rem" }} />;
  }

  return (
    <div
      style={{
        padding: "2rem",
        display: "flex",
        flexDirection: "column",
        gap: "1.5rem",
      }}>
      {showClearConfirm && (
        <ConfirmModal
          title="Clear Blocklist"
          message="Are you sure you want to delete all blocklist entries? This action cannot be undone."
          confirmLabel="Clear All"
          onConfirm={handleClearAll}
          onCancel={() => setShowClearConfirm(false)}
        />
      )}

      {showAddForm && (
        <div
          onClick={() => setShowAddForm(false)}
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
                New Entry
              </h2>
              <button
                type="button"
                onClick={() => setShowAddForm(false)}
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
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "0.75rem",
              }}>
              <div>
                <label style={labelStyle}>Domain</label>
                <input
                  type="text"
                  placeholder="ads.example.com or *.tracking.com"
                  value={domain}
                  onChange={(e) => setDomain(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && domain) {
                      handleAdd();
                      setShowAddForm(false);
                    }
                  }}
                  autoFocus
                  style={inputStyle}
                />
              </div>
              <button
                onClick={() => {
                  handleAdd();
                  setShowAddForm(false);
                }}
                disabled={!domain}
                style={{
                  alignSelf: "flex-end",
                  display: "flex",
                  alignItems: "center",
                  gap: "0.3rem",
                  padding: "0.6rem 1.2rem",
                  background: domain ? colors.primary : colors.border,
                  color: domain ? "#fff" : "#9ca3af",
                  border: "none",
                  borderRadius: "6px",
                  cursor: domain ? "pointer" : "not-allowed",
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

      {showUploadForm && (
        <div
          onClick={() => setShowUploadForm(false)}
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
                <Upload size={18} />
                Upload Blocklist
              </h2>
              <button
                type="button"
                onClick={() => setShowUploadForm(false)}
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
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "0.75rem",
              }}>
              <p
                style={{
                  margin: 0,
                  fontSize: "0.8rem",
                  color: "#64748b",
                  lineHeight: 1.5,
                }}>
                Supports plain domain lists, AdBlock Plus and hosts file
                formats. Max file size 15MB. Duplicates are skipped
                automatically.
              </p>
              <input
                ref={fileRef}
                type="file"
                accept=".txt,.csv,.list"
                style={{
                  fontSize: "0.85rem",
                  fontFamily: "inherit",
                }}
              />
              <button
                onClick={handleUpload}
                disabled={uploading}
                style={{
                  alignSelf: "flex-end",
                  display: "flex",
                  alignItems: "center",
                  gap: "0.3rem",
                  padding: "0.6rem 1.2rem",
                  background: uploading ? colors.border : colors.primary,
                  color: uploading ? "#9ca3af" : "#fff",
                  border: "none",
                  borderRadius: "6px",
                  cursor: uploading ? "not-allowed" : "pointer",
                  fontWeight: 600,
                  fontSize: "0.85rem",
                  fontFamily: "inherit",
                }}>
                {uploading ? (
                  <Loader2
                    size={14}
                    style={{ animation: "spin 1s linear infinite" }}
                  />
                ) : (
                  <Upload size={14} />
                )}
                {uploading ? "Uploading..." : "Upload"}
              </button>
            </div>
          </div>
        </div>
      )}

      <div style={cardStyle}>
        <h2 style={cardTitleStyle}>
          <ShieldBan size={18} />
          Blocklist
          {total > 0 && <Badge>{total.toLocaleString()} entries</Badge>}
        </h2>

        <p
          style={{
            margin: "0 0 0.75rem",
            fontSize: "0.8rem",
            color: "#64748b",
            lineHeight: 1.5,
          }}>
          Blocked domains resolve to 0.0.0.0. Supports wildcards like
          *.ads.example.com
        </p>

        <div
          style={{
            position: "relative",
            marginBottom: "0.75rem",
          }}>
          <Search
            size={14}
            style={{
              position: "absolute",
              left: "0.6rem",
              top: "50%",
              transform: "translateY(-50%)",
              color: "#9ca3af",
            }}
          />
          <input
            type="text"
            placeholder="Search domains..."
            value={search}
            onChange={(e) => handleSearch(e.target.value)}
            style={{
              ...inputStyle,
              paddingLeft: "2rem",
            }}
          />
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
                    <Link size={12} /> Domain
                  </span>
                </th>
                <th style={{ ...thStyle, width: "1%" }}></th>
              </tr>
            </thead>
            <tbody>
              {entries.map((e) => (
                <tr
                  key={e.id}
                  style={{ borderBottom: `1px solid ${colors.border}` }}>
                  <td style={tdStyle}>
                    <Badge mono>{e.domain}</Badge>
                  </td>
                  <td style={tdStyle}>
                    <button
                      onClick={() => handleDelete(e.id)}
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
              {entries.length === 0 && (
                <tr>
                  <td
                    colSpan={2}
                    style={{
                      ...tdStyle,
                      color: "#64748b",
                      textAlign: "center",
                    }}>
                    {search ? "No matching entries" : "No entries"}
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
              {page * perPage + 1}–{Math.min((page + 1) * perPage, total)} of{" "}
              {total.toLocaleString()}
            </span>
            <div style={{ display: "flex", gap: "0.25rem" }}>
              <button
                onClick={() => goToPage(page - 1)}
                disabled={page === 0}
                style={paginationBtnStyle(page === 0)}>
                <ChevronLeft size={14} />
              </button>
              <button
                onClick={() => goToPage(page + 1)}
                disabled={page + 1 >= totalPages}
                style={paginationBtnStyle(page + 1 >= totalPages)}>
                <ChevronRight size={14} />
              </button>
            </div>
          </div>
        )}

        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            gap: "0.5rem",
            marginTop: "0.75rem",
            paddingTop: "0.75rem",
          }}>
          {total > 0 && (
            <button
              onClick={() => setShowClearConfirm(true)}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.3rem",
                padding: "0.6rem 1.2rem",
                background: "transparent",
                color: "#dc2626",
                border: "1px solid #dc2626",
                borderRadius: "6px",
                cursor: "pointer",
                fontWeight: 600,
                fontSize: "0.85rem",
                fontFamily: "inherit",
                marginRight: "auto",
              }}>
              <Trash2 size={14} />
              Clear All
            </button>
          )}
          <button
            onClick={() => setShowUploadForm(true)}
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
              fontWeight: 600,
              fontSize: "0.85rem",
              fontFamily: "inherit",
            }}>
            <Upload size={14} />
            Upload File
          </button>
          <button
            onClick={() => setShowAddForm(true)}
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
            New Entry
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
