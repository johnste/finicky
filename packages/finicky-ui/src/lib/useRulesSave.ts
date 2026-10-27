import { useRef } from "react";
import { api } from "./api";
import { appStore } from "./appStore";
import { toast } from "./toast";
import { useDebouncedCallback } from "./useDebouncedCallback";
import type { RulesFile } from "../types";

// buildPayload is called at flush time (not hook-call time), so callers
// must read their latest data through a ref rather than closing over React
// state directly — otherwise an immediate save() called synchronously
// right after a setState can flush a stale pre-edit value, since React
// hasn't re-rendered yet at that point.
export function useRulesSave(
  buildPayload: () => Partial<RulesFile>,
  delay: number,
  errorMessage = "Failed to save rules"
) {
  const isPending = useRef(false);
  // Saves can overlap (e.g. blur on one field then another), and responses
  // may arrive out of order. Only the latest request's response is applied,
  // so an older one can't overwrite newer state.
  const latestRequest = useRef(0);

  const debounced = useDebouncedCallback(async () => {
    const id = ++latestRequest.current;
    try {
      const updated = await api.saveRules(buildPayload());
      if (id === latestRequest.current) {
        appStore.update({ rulesFile: updated as any });
      }
    } catch (err) {
      toast.error(errorMessage, err instanceof Error ? err.message : String(err));
    } finally {
      if (id === latestRequest.current) {
        isPending.current = false;
      }
    }
  }, delay);

  return {
    isPending,
    save() {
      isPending.current = true;
      debounced.flush();
    },
    scheduleSave() {
      isPending.current = true;
      debounced.schedule();
    },
  };
}
