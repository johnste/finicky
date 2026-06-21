import { useRef } from "react";
import { api } from "./api";
import { appStore } from "./appStore";
import { useDebouncedCallback } from "./useDebouncedCallback";
import type { Rule, RulesFile } from "../types";

export function useRulesSave(rulesFile: RulesFile, getRules: () => Rule[], delay: number) {
  const isPending = useRef(false);

  const debounced = useDebouncedCallback(async () => {
    try {
      const updated = await api.saveRules({
        ...rulesFile,
        rules: getRules(),
      });
      appStore.update({ rulesFile: updated as any });
    } catch {} finally {
      isPending.current = false;
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
