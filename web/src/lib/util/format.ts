// Sizes and dates as people read them (S02.3-T03), in the browser's locale.

const units = ['bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];

/**
 * A size in bytes, in steps of 1,024 as file managers show it:
 * 0 → "0 bytes", 1536 → "1.5 KB", 5_000_000_000 → "4.66 GB".
 */
export function formatSize(bytes: number, locale?: string): string {
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  const digits = unit === 0 ? 0 : value < 10 ? 2 : value < 100 ? 1 : 0;
  const number = new Intl.NumberFormat(locale, { maximumFractionDigits: digits }).format(value);
  return unit === 0 && bytes === 1 ? `${number} byte` : `${number} ${units[unit]}`;
}

/** A modification time: "Sep 25, 2026, 7:15 AM" (the format follows the locale). */
export function formatDate(iso: string, locale?: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return '';
  }
  return new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeStyle: 'short' }).format(date);
}

/** A number with the locale's grouping: 50000 → "50,000". */
export function formatCount(n: number, locale?: string): string {
  return new Intl.NumberFormat(locale).format(n);
}
