import { defineStore } from 'pinia';

export type ToastKind = 'info' | 'success' | 'error' | 'warn';

export interface Toast {
  id: number;
  kind: ToastKind;
  text: string;
  ttl: number;
}

let nextId = 1;

export const useToastStore = defineStore('toast', {
  state: () => ({ items: [] as Toast[] }),
  actions: {
    push(text: string, kind: ToastKind = 'info', ttl = 4000) {
      // De-dup: if the same kind+text is already on screen, skip.
      const dup = this.items.find((t) => t.kind === kind && t.text === text);
      if (dup) return dup.id;
      const id = nextId++;
      this.items.push({ id, kind, text, ttl });
      setTimeout(() => this.dismiss(id), ttl);
      return id;
    },
    info(text: string, ttl?: number) { return this.push(text, 'info', ttl); },
    success(text: string, ttl?: number) { return this.push(text, 'success', ttl); },
    error(text: string, ttl = 6000) { return this.push(text, 'error', ttl); },
    warn(text: string, ttl?: number) { return this.push(text, 'warn', ttl); },
    dismiss(id: number) {
      this.items = this.items.filter((t) => t.id !== id);
    },
  },
});
