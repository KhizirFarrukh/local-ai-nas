import { describe, expect, it } from 'vitest';
import { ApiError, isProblem } from './errors';

describe('ApiError', () => {
  it('takes every field of a problem answer', () => {
    const response = new Response(null, { status: 400, headers: { 'X-Request-ID': 'hdr-id' } });
    const e = ApiError.fromResponse(response, {
      status: 400,
      code: 'invalid_name',
      title: 'Invalid name',
      detail: 'the name "a:b" has a forbidden character',
      rule: 'forbidden_character',
      correlation_id: 'body-id'
    });
    expect(e).toBeInstanceOf(Error);
    expect(e.name).toBe('ApiError');
    expect(e.status).toBe(400);
    expect(e.code).toBe('invalid_name');
    expect(e.rule).toBe('forbidden_character');
    expect(e.message).toBe('the name "a:b" has a forbidden character');
    expect(e.correlationId).toBe('body-id');
  });

  it('uses the title when a problem has no detail, and the header ID when the body has none', () => {
    const response = new Response(null, { status: 404, headers: { 'X-Request-ID': 'hdr-id' } });
    const e = ApiError.fromResponse(response, {
      status: 404,
      code: 'not_found',
      title: 'Not found'
    });
    expect(e.message).toBe('Not found');
    expect(e.correlationId).toBe('hdr-id');
  });

  it('reports an answer that is not a problem', () => {
    const response = new Response('<html>', { status: 502, statusText: 'Bad Gateway' });
    const e = ApiError.fromResponse(response, '<html>');
    expect(e.code).toBe('unexpected_response');
    expect(e.status).toBe(502);
    expect(e.message).toBe('The server answered 502 Bad Gateway.');
    expect(e.correlationId).toBeUndefined();
  });

  it('marks a request that got no answer', () => {
    const cause = new TypeError('Failed to fetch');
    const e = ApiError.unreachable(cause);
    expect(e.status).toBe(0);
    expect(e.code).toBe('unreachable');
    expect(e.cause).toBe(cause);
  });
});

describe('isProblem', () => {
  it('recognizes the problem shape only', () => {
    expect(isProblem({ status: 409, code: 'conflict', title: 'Conflict' })).toBe(true);
    expect(isProblem({ status: '409', code: 'conflict', title: 'Conflict' })).toBe(false);
    expect(isProblem({ code: 'conflict' })).toBe(false);
    expect(isProblem(null)).toBe(false);
    expect(isProblem('conflict')).toBe(false);
  });
});
