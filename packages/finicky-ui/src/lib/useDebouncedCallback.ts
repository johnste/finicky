import { useRef, useEffect } from "react";

export function useDebouncedCallback(fn: () => void | Promise<void>, delay: number) {
  const timer = useRef<ReturnType<typeof setTimeout>>();
  const fnRef = useRef(fn);
  fnRef.current = fn;

  useEffect(() => () => clearTimeout(timer.current), []);

  return {
    schedule() {
      clearTimeout(timer.current);
      timer.current = setTimeout(() => fnRef.current(), delay);
    },
    flush() {
      clearTimeout(timer.current);
      fnRef.current();
    },
  };
}
