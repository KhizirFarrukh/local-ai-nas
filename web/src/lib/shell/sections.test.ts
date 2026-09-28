import { describe, expect, it } from 'vitest';
import { inSection, sections } from './sections';

describe('sections', () => {
  const files = sections[0];
  it('lists Files, Photos, and Settings', () => {
    expect(sections.map((s) => s.label)).toEqual(['Files', 'Photos', 'Settings']);
  });

  it('matches a section and everything below it, not look-alikes', () => {
    expect(inSection('/files', files)).toBe(true);
    expect(inSection('/files/docs/a', files)).toBe(true);
    expect(inSection('/filesystem', files)).toBe(false);
    expect(inSection('/photos', files)).toBe(false);
  });
});
