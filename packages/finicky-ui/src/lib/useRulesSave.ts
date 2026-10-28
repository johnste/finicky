import { useRef } from "react";
import { api } from "./api";
import { appStore } from "./appStore";
import { toast } from "./toast";
import { useDebouncedCallback } from "./useDebouncedCallback";
import type { RulesFile } from "../types";

// Shared by every page that saves the rules file. Each edit is applied to the
// store right away, so a page that remounts (e.g. after switching tabs) starts
// from it rather than from the last saved state, and an edit on one page
// (e.g. Rules) can't be dropped by a save from another (e.g. Preferences).
// Saves are then sent one at a time, in edit order, each with the store's
// rulesFile as it is when its turn comes, so a slow earlier request can't
// overwrite a later one on disk.
let saveQueue: Promise<void> = Promise.resolve();
let queuedSaves = 0;

// buildPatch returns only the fields the calling page owns, and is called when
// the save fires (not at hook-call time), so callers must read their latest data
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
    const save = ++queuedSaves;
    appStore.update({ rulesFile: { ...appStore.getSnapshot().rulesFile, ...buildPatch() } });
    saveQueue = saveQueue.then(async () => {
      // Responses from earlier saves are ignored while a later one is queued,
      // since they don't include the edits that save will send.
      try {
        const updated = await api.saveRules(appStore.getSnapshot().rulesFile);
        if (save === queuedSaves) {
          appStore.update({ rulesFile: updated as any });
        }
      } catch (err) {
        toast.error(errorMessage, err instanceof Error ? err.message : String(err));
        // The store now holds edits that didn't reach disk; reload what's
        // actually saved, unless a later save will send them again.
        if (save === queuedSaves) {
          try {
            appStore.update({ rulesFile: (await api.getRules()) as any });
          } catch {}
        }
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
