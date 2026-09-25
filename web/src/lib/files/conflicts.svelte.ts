// Name conflicts (S02.5-T03, FR-081): when a move, copy, upload, or new
// folder meets a taken name, the user chooses what happens:
// - skip: leave this item;
// - keep both: the server gives the new item a numbered name (`rename`);
// - replace: the new file replaces the old one (`overwrite`). Files only:
//   folders are never merged or replaced (S01 rule), so a folder offers
//   skip and keep both.
// "Apply to all" answers the rest of the operation's conflicts of the same
// kind (file or folder). Operations that run in parallel, as uploads do,
// wait for the question already open instead of asking again.
import type { ConflictChoice } from './operations';

/** What the user is asked about. */
export interface ConflictQuestion {
  /** The name that is taken. */
  name: string;
  /** Whether the new item is a folder: then it cannot replace. */
  folder: boolean;
  /** Where it is taken, as an area path. */
  target: string;
  /** Whether "apply to all" makes sense: the operation has more items. */
  many: boolean;
  /** Whether the operation can be stopped as a whole (bulk runs). */
  stoppable: boolean;
}

export interface ConflictAnswer {
  choice: Exclude<ConflictChoice, 'fail'>;
  all: boolean;
}

interface Asked {
  question: ConflictQuestion;
  resolve: (answer: ConflictAnswer) => void;
}

/** Asks one question at a time; the conflict dialog shows `current`. */
export class ConflictResolver {
  current = $state.raw<Asked | undefined>(undefined);
  private waiting: Asked[] = [];

  ask(question: ConflictQuestion): Promise<ConflictAnswer> {
    return new Promise((resolve) => {
      this.waiting.push({ question, resolve });
      if (!this.current) {
        this.next();
      }
    });
  }

  /** Called by the dialog with the user's answer. */
  answer(answer: ConflictAnswer): void {
    this.current?.resolve(answer);
    this.next();
  }

  private next(): void {
    this.current = this.waiting.shift();
  }
}

/** The conflicts of one operation, which remembers "apply to all". */
export class ConflictBatch {
  private remembered: Partial<Record<'file' | 'folder', ConflictAnswer['choice']>> = {};
  private pending: Partial<Record<'file' | 'folder', Promise<ConflictAnswer>>> = {};

  constructor(
    private readonly resolver: ConflictResolver,
    /** How many items the operation has. */
    private readonly size: number,
    private readonly stoppable = false
  ) {}

  /** The choice for one conflict: remembered, or asked. */
  async choose(name: string, folder: boolean, target: string): Promise<ConflictAnswer['choice']> {
    const kind = folder ? 'folder' : 'file';
    for (;;) {
      const saved = this.remembered[kind];
      if (saved) {
        return saved;
      }
      const open = this.pending[kind];
      if (!open) {
        break;
      }
      await open; // then look again: it may have answered for all
    }
    const asked = this.resolver.ask({
      name,
      folder,
      target,
      many: this.size > 1,
      stoppable: this.stoppable
    });
    this.pending[kind] = asked;
    try {
      const { choice, all } = await asked;
      if (all) {
        this.remembered[kind] = choice;
      }
      return choice;
    } finally {
      if (this.pending[kind] === asked) {
        this.pending[kind] = undefined;
      }
    }
  }
}

/** The app's resolver, shown by the dialog in the layout. */
export const conflicts = new ConflictResolver();
