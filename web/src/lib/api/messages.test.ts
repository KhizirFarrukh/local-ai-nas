import { describe as group, expect, it } from 'vitest';
import { ApiError } from './errors';
import { describe } from './messages';

function err(init: Partial<ConstructorParameters<typeof ApiError>[0]>): ApiError {
  return new ApiError({ status: 400, code: 'invalid_request', message: 'm', ...init });
}

group('describe', () => {
  it('explains each code in plain words', () => {
    expect(describe(err({ code: 'not_found' }))).toEqual({
      title: 'Not found',
      message: 'It may have been moved, renamed, or deleted.',
      detail: undefined
    });
    expect(describe(ApiError.unreachable(new Error('x'))).title).toBe('Can’t reach the NAS');
  });

  it('explains the broken name rule', () => {
    const m = describe(err({ code: 'invalid_name', rule: 'reserved_name' }));
    expect(m.title).toBe('Name not allowed');
    expect(m.message).toMatch(/reserved on Windows/);
    // An unknown rule falls back to the code's message.
    expect(describe(err({ code: 'invalid_name', rule: 'new_rule' })).message).toBe(
      'This name cannot be used.'
    );
  });

  it('passes on the server detail of a conflict and a limit, as a sentence', () => {
    expect(
      describe(err({ code: 'conflict', detail: 'an item already exists at /a' })).message
    ).toBe('An item already exists at /a.');
    expect(
      describe(
        err({ code: 'too_large_for_sync', detail: 'the copy has 20,001 items; at most 20,000.' })
      ).message
    ).toBe('The copy has 20,001 items; at most 20,000. Split it into smaller parts.');
  });

  it('shows the request ID only for errors the server log explains', () => {
    expect(describe(err({ code: 'internal', correlationId: 'abc' })).detail).toBe('Request abc');
    expect(describe(err({ code: 'unavailable', correlationId: 'abc' })).detail).toBe('Request abc');
    expect(describe(err({ code: 'conflict', correlationId: 'abc' })).detail).toBeUndefined();
    expect(describe(err({ code: 'internal' })).detail).toBeUndefined();
  });

  it('describes anything else that was thrown', () => {
    expect(describe(new Error('disk on fire'))).toEqual({
      title: 'Something went wrong',
      message: 'disk on fire'
    });
    expect(describe('plain text').message).toBe('plain text');
  });

  it('falls back for a code it does not know', () => {
    const m = describe(err({ code: 'brand_new' as never, message: 'server words' }));
    expect(m).toEqual({ title: 'Error', message: 'server words', detail: undefined });
  });
});
