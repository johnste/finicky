import { useEffect } from "react";
import { api } from "./api";
import { appStore } from "./appStore";
import type { LogEntry, ConfigInfo, UpdateInfo } from "../types";

export function useSSE() {
  useEffect(() => {
    const es = new EventSource(api.eventsUrl());

    es.addEventListener("log", (e) => {
      try {
        appStore.appendMessage(JSON.parse(e.data) as LogEntry);
      } catch {}
    });

    es.addEventListener("config", (e) => {
      try {
        const config = JSON.parse(e.data) as ConfigInfo;
        appStore.update({ hasConfig: true, config });
      } catch {}
    });

    es.addEventListener("updateInfo", (e) => {
      try {
        appStore.update({ updateInfo: JSON.parse(e.data) as UpdateInfo });
      } catch {}
    });

    return () => es.close();
  }, []);
}
