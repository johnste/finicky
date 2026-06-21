export type ToastType = "success" | "error" | "info" | "warning";

export interface Toast {
  id: string;
  message: string;
  extra?: string;
  type: ToastType;
  duration?: number;
  timeoutId?: number;
  key?: number;
}

const MAX_TOASTS = 3;
const DEFAULT_DURATION = 3000;

type Listener = () => void;

function createToastStore() {
  let toasts: Toast[] = [];
  const listeners = new Set<Listener>();

  function getSnapshot(): Toast[] {
    return toasts;
  }

  function subscribe(listener: Listener): () => void {
    listeners.add(listener);
    return () => listeners.delete(listener);
  }

  function notify() {
    listeners.forEach((l) => l());
  }

  function remove(id: string) {
    const t = toasts.find((t) => t.id === id);
    if (t?.timeoutId) clearTimeout(t.timeoutId);
    toasts = toasts.filter((t) => t.id !== id);
    notify();
  }

  function show(
    message: string,
    type: ToastType = "info",
    extra?: string,
    duration = DEFAULT_DURATION
  ): string {
    const existing = toasts.find((t) => t.message === message && t.type === type);

    if (existing) {
      if (existing.timeoutId) clearTimeout(existing.timeoutId);
      const timeoutId = setTimeout(() => remove(existing.id), duration) as unknown as number;
      toasts = toasts.map((t) =>
        t.id === existing.id ? { ...t, duration, timeoutId, key: Date.now() } : t
      );
      notify();
      return existing.id;
    }

    const id = `${Date.now()}-${Math.random()}`;
    const timeoutId = setTimeout(() => remove(id), duration) as unknown as number;
    const newToast: Toast = { id, message, extra, type, duration, timeoutId, key: Date.now() };

    let updated = [...toasts, newToast];
    if (updated.length > MAX_TOASTS) {
      const oldest = updated[0];
      if (oldest.timeoutId) clearTimeout(oldest.timeoutId);
      updated = updated.slice(-MAX_TOASTS);
    }
    toasts = updated;
    notify();
    return id;
  }

  function clear() {
    toasts.forEach((t) => {
      if (t.timeoutId) clearTimeout(t.timeoutId);
    });
    toasts = [];
    notify();
  }

  return {
    getSnapshot,
    subscribe,
    show,
    success: (message: string, duration?: number) =>
      show(message, "success", undefined, duration),
    error: (message: string, extra?: string, duration?: number) =>
      show(message, "error", extra, duration),
    info: (message: string, extra?: string, duration?: number) =>
      show(message, "info", extra, duration),
    warning: (message: string, duration?: number) =>
      show(message, "warning", undefined, duration),
    remove,
    clear,
  };
}

export const toast = createToastStore();
