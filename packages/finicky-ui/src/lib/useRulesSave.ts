import { useRef } from "react";
import { api } from "./api";
import { appStore } from "./appStore";
import { toast } from "./toast";
import { useDebouncedCallback } from "./useDebouncedCallback";
import type { RulesFile } from "../types";

// Shared by every page that saves the rules file. Saves are sent one at a
// time, in edit order, so a slow earlier request can't overwrite a later one
// on disk. Each request's payload is built when its turn comes, on top of the
// store's rulesFile as updated by the previous save, so a save from one page
// (e.g. Rules) can't drop a change just made on another (e.g. Preferences).
let saveQueue: Promise<void> = Promise.resolve();

// buildPatch returns only the fields the calling page owns, and is called at
// send time (not hook-call time), so callers must read their latest data
// through a ref rather than closing over React state directly — otherwise an
// immediate save() called synchronously right after a setState can send a
// stale pre-edit value, since React hasn't re-rendered yet at that point.
export function useRulesSave(
  buildPatch: () => Partial<RulesFile>,
  delay: number,
  errorMessage = "Failed to save rules"
) {
  // isPending stays set from the first unsaved edit until the save covering
  // the latest edit has finished, so pages don't reset their local state
  // from the store while edits are still on their way to disk.
  const isPending = useRef(false);
  const latestEdit = useRef(0);

  const debounced = useDebouncedCallback(() => {
    const edit = latestEdit.current;
    saveQueue = saveQueue.then(async () => {
      try {
        const updated = await api.saveRules({ ...appStore.getSnapshot().rulesFile, ...buildPatch() });
        appStore.update({ rulesFile: updated as any });
      } catch (err) {
        toast.error(errorMessage, err instanceof Error ? err.message : String(err));
      } finally {
        if (edit === latestEdit.current) {
          isPending.current = false;
        }
      }
    });
  }, delay);

  function markEdited() {
    latestEdit.current++;
    isPending.current = true;
  }

  return {
    isPending,
    save() {
      markEdited();
      debounced.flush();
    },
    scheduleSave() {
      markEdited();
      debounced.schedule();
    },
  };
}
