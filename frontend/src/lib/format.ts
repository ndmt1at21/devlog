// Small presentation helpers shared across components.

/** First grapheme of a name, uppercased — used for avatar monograms. */
export function initial(name: string): string {
  const trimmed = name.trim();
  return trimmed ? trimmed[0].toUpperCase() : "?";
}

/** Format a VND amount as e.g. "75.000đ". */
export function formatVND(amount: number): string {
  return `${amount.toLocaleString("vi-VN")}đ`;
}
