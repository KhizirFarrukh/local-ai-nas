import { describe, expect, it } from 'vitest';
import { ConflictBatch, ConflictResolver } from './conflicts.svelte';

/** Answers the resolver's open question once it appears. */
async function nextQuestion(r: ConflictResolver) {
  for (let i = 0; i < 50 && !r.current; i++) {
    await Promise.resolve();
  }
  return r.current?.question;
}

describe('ConflictResolver', () => {
  it('asks one question at a time, in order', async () => {
    const r = new ConflictResolver();
    const a = r.ask({ name: 'a', folder: false, target: '/a', many: false, stoppable: false });
    const b = r.ask({ name: 'b', folder: true, target: '/b', many: false, stoppable: false });
    expect(r.current?.question.name).toBe('a');
    r.answer({ choice: 'rename', all: false });
    expect(await a).toEqual({ choice: 'rename', all: false });
    expect(r.current?.question.name).toBe('b');
    r.answer({ choice: 'skip', all: false });
    expect(await b).toEqual({ choice: 'skip', all: false });
    expect(r.current).toBeUndefined();
  });
});

describe('ConflictBatch', () => {
  it('asks with the operation size and remembers "apply to all" per kind', async () => {
    const r = new ConflictResolver();
    const batch = new ConflictBatch(r, 3, true);
    const first = batch.choose('a.txt', false, '/to/a.txt');
    expect(r.current?.question).toEqual({
      name: 'a.txt',
      folder: false,
      target: '/to/a.txt',
      many: true,
      stoppable: true
    });
    r.answer({ choice: 'overwrite', all: true });
    expect(await first).toBe('overwrite');
    // Files are answered for all; a folder still asks.
    expect(await batch.choose('b.txt', false, '/to/b.txt')).toBe('overwrite');
    const folder = batch.choose('d', true, '/to/d');
    expect(r.current?.question.folder).toBe(true);
    r.answer({ choice: 'rename', all: false });
    expect(await folder).toBe('rename');
  });

  it('makes parallel conflicts wait for the open question', async () => {
    const r = new ConflictResolver();
    const batch = new ConflictBatch(r, 2);
    const one = batch.choose('a', false, '/a');
    const two = batch.choose('b', false, '/b'); // waits instead of asking at once
    expect((await nextQuestion(r))?.name).toBe('a');
    r.answer({ choice: 'skip', all: true });
    expect(await one).toBe('skip');
    expect(await two).toBe('skip');
    expect(r.current).toBeUndefined();
  });

  it('asks the waiting conflict itself when the first answer was not for all', async () => {
    const r = new ConflictResolver();
    const batch = new ConflictBatch(r, 2);
    const one = batch.choose('a', false, '/a');
    const two = batch.choose('b', false, '/b');
    r.answer({ choice: 'rename', all: false });
    expect(await one).toBe('rename');
    expect((await nextQuestion(r))?.name).toBe('b');
    expect(r.current?.question.many).toBe(true);
    r.answer({ choice: 'overwrite', all: false });
    expect(await two).toBe('overwrite');
  });

  it('offers no "apply to all" for a single item', () => {
    const r = new ConflictResolver();
    void new ConflictBatch(r, 1).choose('a', false, '/a');
    expect(r.current?.question.many).toBe(false);
    expect(r.current?.question.stoppable).toBe(false);
  });
});
