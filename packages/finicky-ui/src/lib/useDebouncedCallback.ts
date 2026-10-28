import { useRef, useEffect } from "react";

// A call that is still scheduled when the component unmounts runs right away
// instead of being dropped, so e.g. an edit followed by switching tabs within
// the debounce delay is still saved.
export function useDebouncedCallback(fn: () => void | Promise<void>, delay: number) {
  const timer = useRef<ReturnType<typeof setTimeout>>();
  const scheduled = useRef(false);
  const fnRef = useRef(fn);
  fnRef.current = fn;

  useEffect(
    () => () => {
      clearTimeout(timer.current);
      if (scheduled.current) {
        scheduled.current = false;
        fnRef.current();
      }
    },
    []
  );

  return {
    schedule() {
      clearTimeout(timer.current);
      scheduled.current = true;
      timer.current = setTimeout(() => {
        scheduled.current = false;
        fnRef.current();
      }, delay);
    },
    flush() {
      clearTimeout(timer.current);
      scheduled.current = false;
      fnRef.current();
    },
  };
}
