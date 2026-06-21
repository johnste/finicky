import { useRef } from "react";
import { api } from "./api";
import { appStore } from "./appStore";
import { useDebouncedCallback } from "./useDebouncedCallback";
import type { Rule, RulesFile } from "../types";

export function useRulesSave(rulesFile: RulesFile, getRules: () => Rule[], delay: number) {
  const isPending = useRef(false);

  const debounced = useDebouncedCallback(async () => {
    isPending.current = false;
    try {
      const updated = await api.saveRules({
        defaultBrowser: rulesFile.defaultBrowser,
        defaultProfile: rulesFile.defaultProfile,
        rules: getRules(),
      });
      appStore.update({ rulesFile: updated as any });
    } catch {}
  }, delay);

  return {
    isPending,
    save() { debounced.flush(); },
    scheduleSave() {
      isPending.current = true;
      debounced.schedule();
    },
  };
}
