// API errors (S02.1-T03). Every error answer of the API is an RFC 9457
// problem (docs/api/errors.md). ApiError carries its fields, so screens can
// react to the stable `code` and `rule` and show the request ID.
import type { components } from './schema';

export type Problem = components['schemas']['Problem'];
export type ProblemCode = Problem['code'];

/** Codes the client adds for failures that never reached an API answer. */
export type ClientCode = 'unreachable' | 'unexpected_response';

export class ApiError extends Error {
  /** The HTTP status; 0 when the server could not be reached. */
  readonly status: number;
  readonly code: ProblemCode | ClientCode;
  /** The broken name rule, for invalid_name (errors.md, Name rules). */
  readonly rule?: string;
  readonly detail?: string;
  /** The request ID: the same ID is in the server log. */
  readonly correlationId?: string;

  constructor(init: {
    status: number;
    code: ProblemCode | ClientCode;
    message: string;
    rule?: string;
    detail?: string;
    correlationId?: string;
    cause?: unknown;
  }) {
    super(init.message, { cause: init.cause });
    this.name = 'ApiError';
    this.status = init.status;
    this.code = init.code;
    this.rule = init.rule;
    this.detail = init.detail;
    this.correlationId = init.correlationId;
  }

  /** An error for a request that got no answer at all. */
  static unreachable(cause: unknown): ApiError {
    return new ApiError({
      status: 0,
      code: 'unreachable',
      message: 'The server could not be reached.',
      cause
    });
  }

  /** An error for an answer that is not a success. */
  static fromResponse(response: Response, body: unknown): ApiError {
    const correlationId = response.headers.get('X-Request-ID') ?? undefined;
    if (isProblem(body)) {
      return new ApiError({
        status: body.status,
        code: body.code,
        message: body.detail ?? body.title,
        rule: body.rule,
        detail: body.detail,
        correlationId: body.correlation_id ?? correlationId
      });
    }
    return new ApiError({
      status: response.status,
      code: 'unexpected_response',
      message: `The server answered ${response.status} ${response.statusText}.`.trim(),
      correlationId
    });
  }
}

/** Reports whether a value has the shape of an API problem. */
export function isProblem(value: unknown): value is Problem {
  if (typeof value !== 'object' || value === null) {
    return false;
  }
  const p = value as Record<string, unknown>;
  return typeof p.status === 'number' && typeof p.code === 'string' && typeof p.title === 'string';
}
