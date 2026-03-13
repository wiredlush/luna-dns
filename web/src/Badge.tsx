interface BadgeProps {
  children: React.ReactNode;
  color?: string;
  mono?: boolean;
}

const defaultColor = "#64748b";

export default function Badge({ children, color, mono }: BadgeProps) {
  const c = color || defaultColor;
  const isDefault = !color;

  return (
    <span
      style={{
        fontSize: "0.7rem",
        color: c,
        background: isDefault ? "#f1f5f9" : c + "14",
        border: `1px solid ${isDefault ? "#e2e8f0" : c}`,
        borderRadius: "9999px",
        padding: "0.15rem 0.5rem",
        fontWeight: 600,
        whiteSpace: "nowrap",
        fontFamily: mono ? "monospace" : undefined,
      }}>
      {children}
    </span>
  );
}
