"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

const ALL = "Tất cả";

/**
 * Category filter as a single-row, horizontally-scrollable chip bar. Instead of
 * wrapping into many rows (which grows unbounded as categories are added), the
 * pills stay on one line and scroll: edge fades hint at more, desktop arrow
 * buttons appear only when there's overflow, and the active pill is scrolled
 * into view so a URL-restored filter is always visible.
 */
export function CategoryFilter({
  categories,
  active,
  onSelect,
}: {
  categories: string[];
  active: string;
  onSelect: (cat: string) => void;
}) {
  const t = useT();
  const scrollerRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [edges, setEdges] = useState({ left: false, right: false });

  const updateEdges = useCallback(() => {
    const el = scrollerRef.current;
    if (!el) return;
    const { scrollLeft, scrollWidth, clientWidth } = el;
    setEdges({
      left: scrollLeft > 1,
      right: scrollLeft + clientWidth < scrollWidth - 1,
    });
  }, []);

  // Recompute the overflow state on scroll and whenever the viewport or the
  // chip row's own width changes (resize, font load, category list changes).
  useEffect(() => {
    const el = scrollerRef.current;
    if (!el) return;
    updateEdges();
    el.addEventListener("scroll", updateEdges, { passive: true });
    const ro = new ResizeObserver(updateEdges);
    ro.observe(el);
    if (contentRef.current) ro.observe(contentRef.current);
    return () => {
      el.removeEventListener("scroll", updateEdges);
      ro.disconnect();
    };
  }, [updateEdges]);

  // Bring the active chip into view (e.g. after a reload restores a filter that
  // sits off-screen). Uses viewport rects so it's independent of the offset
  // parent, and scrollBy so only the row moves — never the page.
  useEffect(() => {
    const el = scrollerRef.current;
    const chip = el?.querySelector<HTMLElement>('[data-active="true"]');
    if (!el || !chip) return;
    const elRect = el.getBoundingClientRect();
    const chipRect = chip.getBoundingClientRect();
    if (chipRect.left < elRect.left || chipRect.right > elRect.right) {
      const delta =
        chipRect.left - elRect.left - (el.clientWidth - chipRect.width) / 2;
      el.scrollBy({ left: delta, behavior: "smooth" });
    }
  }, [active]);

  const nudge = (dir: -1 | 1) => {
    const el = scrollerRef.current;
    if (!el) return;
    el.scrollBy({ left: dir * el.clientWidth * 0.8, behavior: "smooth" });
  };

  return (
    <div className="relative mx-auto mt-11 max-w-[1120px]">
      <div
        ref={scrollerRef}
        role="group"
        aria-label={t("common.allCategories")}
        className="overflow-x-auto px-6 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      >
        <div ref={contentRef} className="flex w-max gap-[9px] py-0.5">
          {categories.map((cat) => {
            const isActive = cat === active;
            return (
              <button
                key={cat}
                data-active={isActive}
                onClick={() => {
                  onSelect(cat);
                  track("select_category", { category: cat });
                }}
                aria-pressed={isActive}
                className="shrink-0 cursor-pointer rounded-full px-4 py-2 text-[14px] font-semibold transition-[filter] hover:brightness-[.98]"
                style={
                  isActive
                    ? {
                        background: "var(--pill-bg)",
                        color: "var(--pill-text)",
                        border: "1px solid var(--pill-bg)",
                      }
                    : {
                        background: "var(--surface)",
                        color: "var(--c43)",
                        border: "1px solid var(--border-2)",
                      }
                }
              >
                {cat === ALL ? t("common.allCategories") : cat}
              </button>
            );
          })}
        </div>
      </div>

      {/* Left edge: fade mask + (desktop) scroll-back arrow. */}
      {edges.left && (
        <>
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-y-0 left-6 w-14 bg-gradient-to-r from-[var(--bg)] to-transparent"
          />
          <button
            type="button"
            aria-label={t("home.scrollCategoriesLeft")}
            onClick={() => nudge(-1)}
            className="absolute left-5 top-1/2 hidden h-9 w-9 -translate-y-1/2 place-items-center rounded-full border border-border2 bg-surface text-[18px] leading-none text-c5b shadow-[0_2px_12px_-4px_rgba(0,0,0,.28)] transition-colors hover:text-strong sm:grid"
          >
            ‹
          </button>
        </>
      )}

      {/* Right edge: fade mask + (desktop) scroll-forward arrow. */}
      {edges.right && (
        <>
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-y-0 right-6 w-14 bg-gradient-to-l from-[var(--bg)] to-transparent"
          />
          <button
            type="button"
            aria-label={t("home.scrollCategoriesRight")}
            onClick={() => nudge(1)}
            className="absolute right-5 top-1/2 hidden h-9 w-9 -translate-y-1/2 place-items-center rounded-full border border-border2 bg-surface text-[18px] leading-none text-c5b shadow-[0_2px_12px_-4px_rgba(0,0,0,.28)] transition-colors hover:text-strong sm:grid"
          >
            ›
          </button>
        </>
      )}
    </div>
  );
}
