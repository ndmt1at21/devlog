"use client";

import { useEffect, useRef } from "react";

/**
 * Fires `onImpression` once when the element first becomes visible (>= threshold
 * of it on screen). Reusable for ads or any view-tracked element.
 */
export function useImpression<T extends HTMLElement>(
  onImpression: () => void,
  threshold = 0.5,
) {
  const ref = useRef<T>(null);
  const fired = useRef(false);

  useEffect(() => {
    const el = ref.current;
    if (!el || fired.current) return;
    if (typeof IntersectionObserver === "undefined") return;

    const obs = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting && !fired.current) {
            fired.current = true;
            onImpression();
            obs.disconnect();
          }
        }
      },
      { threshold },
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [onImpression, threshold]);

  return ref;
}
