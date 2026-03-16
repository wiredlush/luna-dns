import { Trash2, X } from "lucide-react";
import { colors } from "./theme";

interface ConfirmModalProps {
  title: string;
  message: React.ReactNode;
  confirmLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export default function ConfirmModal({
  title,
  message,
  confirmLabel = "Delete",
  onConfirm,
  onCancel,
}: ConfirmModalProps) {
  return (
    <div
      onClick={onCancel}
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
            {title}
          </h2>
          <button
            type="button"
            onClick={onCancel}
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
        <p style={{ margin: "0 0 1.25rem", fontSize: "0.9rem" }}>{message}</p>
        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            gap: "0.5rem",
          }}>
          <button
            onClick={onCancel}
            style={{
              padding: "0.6rem 1.2rem",
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
            onClick={onConfirm}
            style={{
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
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
