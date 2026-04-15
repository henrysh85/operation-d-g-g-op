import { defineStore } from 'pinia';
import type { Region, Vertical } from '@/types';

interface State {
  vertical: Vertical | 'all';
  region:   Region   | 'all';
  clientId: string   | 'all';
  month:    string;   // YYYY-MM
  search:   string;
  debouncedSearch: string;
}

function currentMonth(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

export const useFiltersStore = defineStore('filters', {
  state: (): State => ({
    vertical: 'all',
    region:   'all',
    clientId: 'all',
    month:    currentMonth(),
    search:   '',
    debouncedSearch: '',
  }),
  getters: {
    // Downstream watchers read asQuery, which uses the debounced search so
    // every keystroke doesn't fire a round-trip.
    asQuery(state): Record<string, string> {
      const q: Record<string, string> = {};
      if (state.vertical !== 'all') q.vertical = state.vertical;
      if (state.region   !== 'all') q.region   = state.region;
      if (state.clientId !== 'all') q.clientId = state.clientId;
      if (state.debouncedSearch)    q.q        = state.debouncedSearch;
      // month (YYYY-MM) fans out to from / to so endpoints that accept a date
      // range (e.g. activities) actually filter, while others that don't look
      // at either just ignore them.
      if (state.month && /^\d{4}-\d{2}$/.test(state.month)) {
        const [y, m] = state.month.split('-').map(Number);
        const last = new Date(y, m, 0).getDate(); // day 0 of next month = last of this
        q.from = `${state.month}-01`;
        q.to   = `${state.month}-${String(last).padStart(2, '0')}`;
        q.month = state.month;
      }
      return q;
    },
  },
  actions: {
    set<K extends keyof State>(key: K, value: State[K]) {
      (this as State)[key] = value;
    },
    reset() {
      this.vertical = 'all';
      this.region   = 'all';
      this.clientId = 'all';
      this.month    = currentMonth();
      this.search   = '';
      this.debouncedSearch = '';
    },
    hydrateFromQuery(q: Record<string, string | string[] | undefined>) {
      const pick = (k: string) => (typeof q[k] === 'string' ? (q[k] as string) : undefined);
      this.vertical = (pick('vertical') as State['vertical']) ?? this.vertical;
      this.region   = (pick('region')   as State['region'])   ?? this.region;
      this.clientId = pick('clientId') ?? this.clientId;
      this.month    = pick('month')    ?? this.month;
      this.search   = pick('q')        ?? this.search;
    },
  },
});
