import { describe, expect, it } from 'vitest';
import { formatCount, formatDate, formatSize } from './format';

describe('formatSize', () => {
  it('uses steps of 1,024 with fewer decimals as numbers grow', () => {
    expect(formatSize(0, 'en-US')).toBe('0 bytes');
    expect(formatSize(1, 'en-US')).toBe('1 byte');
    expect(formatSize(1023, 'en-US')).toBe('1,023 bytes');
    expect(formatSize(1536, 'en-US')).toBe('1.5 KB');
    expect(formatSize(20_000, 'en-US')).toBe('19.5 KB');
    expect(formatSize(5_000_000_000, 'en-US')).toBe('4.66 GB');
    expect(formatSize(300 * 1024 ** 3, 'en-US')).toBe('300 GB');
  });

  it('stops at petabytes', () => {
    expect(formatSize(2048 * 1024 ** 5, 'en-US')).toBe('2,048 PB');
  });

  it('follows the locale', () => {
    expect(formatSize(1536, 'de-DE')).toBe('1,5 KB');
  });
});

describe('formatDate', () => {
  it('shows a medium date with a short time', () => {
    const got = formatDate('2026-09-25T07:15:00Z', 'en-US');
    expect(got).toMatch(/Sep 2[45], 2026/);
    expect(got).toMatch(/\d{1,2}:15/);
  });

  it('gives nothing for an invalid time', () => {
    expect(formatDate('not a time', 'en-US')).toBe('');
  });
});

describe('formatCount', () => {
  it('groups digits', () => {
    expect(formatCount(50_000, 'en-US')).toBe('50,000');
    expect(formatCount(7, 'en-US')).toBe('7');
  });
});
