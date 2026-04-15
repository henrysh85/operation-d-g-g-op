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
      if (state.month)              q.month    = state.month;
      if (state.debouncedSearch)    q.q        = state.debouncedSearch;
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
