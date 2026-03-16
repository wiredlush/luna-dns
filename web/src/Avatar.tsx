import { colors } from "./theme";

export default function Avatar({
  name,
  size = 28,
}: {
  name: string;
  size?: number;
}) {
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        width: size,
        height: size,
        borderRadius: "50%",
        background: colors.primary,
        color: "#fff",
        fontSize: "0.75rem",
        fontWeight: 700,
        flexShrink: 0,
      }}>
      {name.charAt(0).toUpperCase()}
    </span>
  );
}
