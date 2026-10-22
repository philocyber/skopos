"use client";

import * as React from "react";

// usePolling runs `fn` on an interval, but pauses while the tab is hidden and
// resumes (with an immediate run) when it becomes visible again. A long-running
// console is often left open in a background tab; raw setInterval keeps hitting
// the backend every few seconds regardless, which this avoids.
//
// `fn` is kept in a ref so the interval does not reset when a new closure is
// passed each render; list `enabled` (and anything that should restart polling)
// in `deps`.
export function usePolling(fn: () => void, intervalMs: number, deps: React.DependencyList = [], enabled = true) {
  const saved = React.useRef(fn);
  saved.current = fn;

  // biome-ignore lint/correctness/useExhaustiveDependencies: deps are the caller's explicit restart triggers.
  React.useEffect(() => {
    if (!enabled) return;
    let timer: ReturnType<typeof setInterval> | undefined;

    const tick = () => saved.current();
    const start = () => {
      if (timer === undefined) timer = setInterval(tick, intervalMs);
    };
    const stop = () => {
      if (timer !== undefined) {
        clearInterval(timer);
        timer = undefined;
      }
    };
    const onVisibility = () => {
      if (typeof document !== "undefined" && document.hidden) {
        stop();
      } else {
        tick(); // catch up immediately on resume
        start();
      }
    };

    tick();
    if (typeof document === "undefined" || !document.hidden) start();
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      document.removeEventListener("visibilitychange", onVisibility);
      stop();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [intervalMs, enabled, ...deps]);
}
