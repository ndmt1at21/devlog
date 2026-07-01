"use client";

import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";

/** Thin reading-progress bar shown on article pages (mockup: top:67px accent bar). */
export function ScrollProgress() {
  const pathname = usePathname();
  const [pct, setPct] = useState(0);
  const active = pathname.startsWith("/articles/");

  useEffect(() => {
    if (!active) return;
    const onScroll = () => {
      const h = document.documentElement;
      const max = h.scrollHeight - h.clientHeight;
      setPct(max > 0 ? Math.min(100, (h.scrollTop / max) * 100) : 0);
    };
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    window.addEventListener("resize", onScroll);
    return () => {
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("resize", onScroll);
    };
  }, [active, pathname]);

  if (!active) return null;
  return (
    <div
      aria-hidden="true"
      className="fixed left-0 top-[67px] z-30 h-[3px] rounded-r-[2px] bg-accent transition-[width] duration-75 ease-linear"
      style={{ width: `${pct}%` }}
    />
  );
}
