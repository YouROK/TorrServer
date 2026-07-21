/** CSS color-mix helper (replaces polished rgba). */
export const alphaCss = (color: string, alpha: number) =>
  `color-mix(in srgb, ${color} ${Math.round(alpha * 100)}%, transparent)`
