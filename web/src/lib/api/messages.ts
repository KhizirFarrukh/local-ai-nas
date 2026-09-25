// Plain-language messages for API errors (S02.2-T02, FR-079): what went
// wrong and what to do, for people who never read docs/api/errors.md.
// Screens show `title` and `message`; `detail` carries the request ID for
// errors worth reporting.
import { ApiError, type ClientCode, type ProblemCode } from './errors';

export interface ErrorMessage {
  title: string;
  message: string;
  /** For example "Request 3f9c…", to find the server log entry. */
  detail?: string;
}

const byCode: Record<ProblemCode | ClientCode, Omit<ErrorMessage, 'detail'>> = {
  unreachable: {
    title: 'Can’t reach the NAS',
    message: 'Check that the local-ai-nas server is running on this computer.'
  },
  unexpected_response: {
    title: 'Unexpected answer',
    message: 'The server gave an answer the app does not understand. Try again.'
  },
  internal: {
    title: 'Something went wrong on the NAS',
    message: 'The server ran into a problem. The request ID below points to the server log.'
  },
  invalid_request: { title: 'Not accepted', message: 'The server did not accept the request.' },
  invalid_name: { title: 'Name not allowed', message: 'This name cannot be used.' },
  outside_root: {
    title: 'Outside your files',
    message: 'That location is outside your files, or goes through a link.'
  },
  not_found: { title: 'Not found', message: 'It may have been moved, renamed, or deleted.' },
  conflict: { title: 'Already there', message: 'Something with this name is already there.' },
  too_large: { title: 'Too large', message: 'This is larger than the server allows.' },
  insufficient_storage: {
    title: 'Not enough space',
    message: 'The NAS does not have enough free space for this.'
  },
  not_available: { title: 'Not available yet', message: 'This comes in a later version.' },
  method_not_allowed: { title: 'Not allowed', message: 'The server does not allow this action.' },
  length_required: {
    title: 'Size unknown',
    message: 'The size of the upload must be known in advance.'
  },
  precondition_failed: {
    title: 'Changed meanwhile',
    message: 'The file changed while this was going on. Try again.'
  },
  range_not_satisfiable: {
    title: 'Outside the file',
    message: 'The requested part lies outside the file.'
  },
  too_large_for_sync: {
    title: 'Too much at once',
    message: 'Copy at most 1,000 items or 1 GiB at a time. Split the copy into smaller parts.'
  },
  locked: { title: 'Busy', message: 'This is being changed right now. Try again in a moment.' },
  unavailable: {
    title: 'The NAS is busy',
    message: 'The server cannot take this right now. Try again in a moment.'
  }
};

/** Explanations of the name rules (docs/api/errors.md, Name rules). */
const byRule: Record<string, string> = {
  empty_name: 'A name cannot be empty.',
  dot_name: '“.” and “..” cannot be used as names.',
  reserved_name:
    'This name is reserved on Windows (for example CON, NUL, AUX, or COM1, with any extension).',
  forbidden_character: 'Names cannot contain < > : " / \\ | ? or *.',
  control_character: 'Names cannot contain control characters.',
  trailing_dot_or_space: 'Names cannot end with a dot or a space.',
  name_too_long: 'The name is too long (at most 255 bytes).',
  path_too_long: 'The whole path is too long.',
  invalid_utf8: 'The name contains characters that are not valid text.',
  lookalike_separator: 'Names cannot contain characters that look like a slash.'
};

/** Codes whose request ID is worth showing: the server log explains them. */
const reportable = new Set<string>(['internal', 'unexpected_response', 'unavailable']);

/** Starts a server detail with a capital letter and ends it with a full stop. */
function sentence(text: string): string {
  const t = text.charAt(0).toUpperCase() + text.slice(1);
  return /[.!?]$/.test(t) ? t : t + '.';
}

/** Describes any thrown value for the user. */
export function describe(error: unknown): ErrorMessage {
  if (!(error instanceof ApiError)) {
    return {
      title: 'Something went wrong',
      message: error instanceof Error ? error.message : String(error)
    };
  }
  const base = byCode[error.code] ?? { title: 'Error', message: error.message };
  const message =
    error.code === 'invalid_name' && error.rule && byRule[error.rule]
      ? byRule[error.rule]
      : error.code === 'conflict' && error.detail
        ? sentence(error.detail)
        : base.message;
  const detail =
    reportable.has(error.code) && error.correlationId
      ? `Request ${error.correlationId}`
      : undefined;
  return { title: base.title, message, detail };
}
