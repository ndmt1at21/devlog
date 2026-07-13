"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { ArticleSummary } from "@/lib/types";
import { useT } from "@/lib/i18n/provider";
import { FeaturedCard } from "./FeaturedCard";

const AUTOPLAY_MS = 6000;
// Minimum horizontal swipe distance (px) to change slide on touch.
const SWIPE_PX = 40;

function ArrowIcon({ dir }: { dir: "prev" | "next" }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="18"
      height="18"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {dir === "prev" ? <path d="M15 18l-6-6 6-6" /> : <path d="M9 6l6 6-6 6" />}
    </svg>
  );
}

/**
 * One-at-a-time slider for the featured articles. Slides advance on arrows,
 * dots, touch swipe, or a gentle autoplay that pauses while hovered/focused.
 * Off-screen slides are `inert` so their links don't catch keyboard focus.
 */
export function FeaturedCarousel({ articles }: { articles: ArticleSummary[] }) {
  const t = useT();
  const count = articles.length;
  const [index, setIndex] = useState(0);
  const [paused, setPaused] = useState(false);
  const touchX = useRef<number | null>(null);

  const go = useCallback(
    (i: number) => setIndex(((i % count) + count) % count),
    [count],
  );

  useEffect(() => {
    if (paused || count <= 1) return;
    const id = window.setInterval(
      () => setIndex((i) => (i + 1) % count),
      AUTOPLAY_MS,
    );
    return () => window.clearInterval(id);
  }, [paused, count]);

  if (count === 0) return null;
  if (count === 1) return <FeaturedCard article={articles[0]} />;

  const arrowBtn =
    "absolute top-1/2 z-[1] hidden h-9 w-9 -translate-y-1/2 items-center justify-center rounded-full border border-border bg-surface text-c3a shadow-[0_6px_18px_-8px_rgba(0,0,0,.25)] transition-colors hover:border-accent-ink hover:text-accent-ink sm:flex";

  return (
    <div
      role="group"
      aria-roledescription="carousel"
      aria-label={t("home.featured")}
      className="relative"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
      onFocusCapture={() => setPaused(true)}
      onBlurCapture={() => setPaused(false)}
      onTouchStart={(e) => {
        touchX.current = e.touches[0].clientX;
      }}
      onTouchEnd={(e) => {
        if (touchX.current === null) return;
        const dx = e.changedTouches[0].clientX - touchX.current;
        touchX.current = null;
        if (Math.abs(dx) >= SWIPE_PX) go(index + (dx < 0 ? 1 : -1));
      }}
    >
      <div className="overflow-hidden rounded-[18px]">
        <div
          className="flex transition-transform duration-500 ease-out motion-reduce:transition-none"
          style={{ transform: `translateX(-${index * 100}%)` }}
        >
          {articles.map((a, i) => (
            <div
              key={a.slug}
              className="w-full flex-none"
              aria-hidden={i !== index}
              inert={i !== index}
            >
              <FeaturedCard article={a} position={i} />
            </div>
          ))}
        </div>
      </div>

      <button
        type="button"
        onClick={() => go(index - 1)}
        aria-label={t("home.carouselPrev")}
        className={`${arrowBtn} left-3`}
      >
        <ArrowIcon dir="prev" />
      </button>
      <button
        type="button"
        onClick={() => go(index + 1)}
        aria-label={t("home.carouselNext")}
        className={`${arrowBtn} right-3`}
      >
        <ArrowIcon dir="next" />
      </button>

      <div className="mt-4 flex items-center justify-center gap-2">
        {articles.map((a, i) => (
          <button
            key={a.slug}
            type="button"
            onClick={() => go(i)}
            aria-label={t("home.carouselGoTo", { n: i + 1 })}
            aria-current={i === index}
            className={`h-2 rounded-full transition-all duration-300 ${
              i === index
                ? "w-5 bg-accent"
                : "w-2 bg-border hover:bg-hover"
            }`}
          />
        ))}
      </div>
    </div>
  );
}
