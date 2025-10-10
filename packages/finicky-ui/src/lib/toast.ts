import { writable } from "svelte/store";

export type ToastType = "success" | "error" | "info" | "warning";

export interface Toast {
  id: string;
  message: string;
  extra?: string;
  type: ToastType;
  duration?: number;
  timeoutId?: number;
  key?: number; // Used to force animation restart
}

const MAX_TOASTS = 3;
const DEFAULT_DURATION = 3000;

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([]);

  function show(message: string, type: ToastType = "info", extra?: string, duration = DEFAULT_DURATION) {
    let id: string;
    let timeoutId: number;

    update((toasts) => {
      // Check if toast with same message and type already exists
      const existing = toasts.find((t) => t.message === message && t.type === type);

      if (existing) {
        // Clear existing timeout
        if (existing.timeoutId) {
          clearTimeout(existing.timeoutId);
        }
        id = existing.id;

        // Reset timeout
        timeoutId = setTimeout(() => {
          remove(id);
        }, duration) as unknown as number;

        // Update the existing toast with new timeout and key to force animation restart
        return toasts.map((t) =>
          t.id === existing.id ? { ...t, duration, timeoutId, key: Date.now() } : t
        );
      } else {
        // Create new toast
        id = `${Date.now()}-${Math.random()}`;

        timeoutId = setTimeout(() => {
          remove(id);
        }, duration) as unknown as number;

        const newToast: Toast = { id, message, extra, type, duration, timeoutId, key: Date.now() };
        const updatedToasts = [...toasts, newToast];

        // Limit max toasts (remove oldest)
        if (updatedToasts.length > MAX_TOASTS) {
          const removed = updatedToasts[0];
          if (removed.timeoutId) {
            clearTimeout(removed.timeoutId);
          }
          return updatedToasts.slice(-MAX_TOASTS);
        }

        return updatedToasts;
      }
    });

    return id!;
  }

  function remove(id: string) {
    update((toasts) => {
      const toast = toasts.find((t) => t.id === id);
      if (toast?.timeoutId) {
        clearTimeout(toast.timeoutId);
      }
      return toasts.filter((t) => t.id !== id);
    });
  }

  function clear() {
    update((toasts) => {
      toasts.forEach((t) => {
        if (t.timeoutId) {
          clearTimeout(t.timeoutId);
        }
      });
      return [];
    });
  }

  return {
    subscribe,
    show,
    success: (message: string, duration?: number) => show(message, "success", duration),
    error: (message: string, error: string, duration?: number) => show(message, "error", error, duration),
    info: (message: string, duration?: number) => show(message, "info", duration),
    warning: (message: string, duration?: number) => show(message, "warning", duration),
    remove,
    clear,
  };
}

export const toast = createToastStore();
