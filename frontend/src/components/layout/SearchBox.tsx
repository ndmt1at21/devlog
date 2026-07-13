"use client";

import { useEffect, useId, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import type { ArticleSummary } from "@/lib/types";
import { api } from "@/lib/api";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

// Wait for a typing pause before hitting the API.
const DEBOUNCE_MS = 300;
// Skip the API for ultra-short queries that would match everything.
const MIN_CHARS = 2;
// The dropdown shows at most this many picks; "view all" leads to the full list.
const MAX_SHOWN = 8;

/**
 * Header search: a combobox that queries the backend (GET /articles?q=) after a
 * typing pause and drops down a pickable result list (title, excerpt, meta).
 * Enter without a highlighted item — or "view all" — goes to the homepage
 * results view at /?q=…
 */
export function SearchBox({
  large = false,
  autoFocus = false,
  onNavigate,
}: {
  /** Mobile-overlay sizing. */
  large?: boolean;
  autoFocus?: boolean;
  /** Called right before navigating (e.g. to close the mobile overlay). */
  onNavigate?: () => void;
}) {
  const t = useT();
  const router = useRouter();
  const listId = useId();
  const wrapRef = useRef<HTMLDivElement>(null);
  const seq = useRef(0);

  const [query, setQuery] = useState("");
  const [results, setResults] = useState<ArticleSummary[] | null>(null);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(-1);
  const [loading, setLoading] = useState(false);

  const q = query.trim();
  const shown = (results ?? []).slice(0, MAX_SHOWN);

  // Short queries reset synchronously in the change handler (not in the effect,
  // which must stay async-only); longer ones flip the loading flag immediately.
  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    if (value.trim().length < MIN_CHARS) {
      setResults(null);
      setLoading(false);
      setOpen(false);
      setActive(-1);
    } else {
      setLoading(true);
    }
  };

  // Debounced API search; a sequence counter discards out-of-order responses.
  useEffect(() => {
    if (q.length < MIN_CHARS) return;
    const timer = window.setTimeout(async () => {
      const id = ++seq.current;
      try {
        const list = await api.searchArticles(q);
        if (seq.current !== id) return;
        setResults(list);
        setActive(-1);
        setOpen(true);
        track("search", { search_term: q, results_count: list.length });
      } catch {
        if (seq.current !== id) return;
        setResults([]);
        setOpen(true);
      } finally {
        if (seq.current === id) setLoading(false);
      }
    }, DEBOUNCE_MS);
    return () => window.clearTimeout(timer);
  }, [q]);

  // Close when clicking/tapping outside (same pattern as AccountMenu).
  useEffect(() => {
    if (!open) return;
    const onPointerDown = (e: PointerEvent) => {
      if (!wrapRef.current?.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [open]);

  const pick = (a: ArticleSummary, position: number) => {
    track("select_article", {
      slug: a.slug,
      title: a.title,
      category: a.category,
      list: "search-suggest",
      position,
    });
    setOpen(false);
    onNavigate?.();
    router.push(`/articles/${encodeURIComponent(a.slug)}`);
  };

  const viewAll = () => {
    setOpen(false);
    onNavigate?.();
    // Carry an active category filter (if any) through to the results view so
    // the URL keeps the full state. Read it from the current URL at click time
    // to avoid useSearchParams, which would opt static pages into client render.
    const params = new URLSearchParams();
    params.set("q", q);
    const category =
      typeof window !== "undefined"
        ? new URLSearchParams(window.location.search).get("category")
        : null;
    if (category) params.set("category", category);
    router.push(`/?${params.toString()}`);
  };

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Escape") {
      setOpen(false);
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      if (open && active >= 0 && shown[active]) pick(shown[active], active);
      else if (q.length >= MIN_CHARS) viewAll();
      return;
    }
    if (!shown.length) return;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setOpen(true);
      setActive((i) => (i + 1) % shown.length);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setOpen(true);
      setActive((i) => (i - 1 + shown.length) % shown.length);
    }
  };

  return (
    <div ref={wrapRef} className="relative w-full">
      <span
        aria-hidden="true"
        className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[15px] text-faint"
      >
        ⌕
      </span>
      <input
        value={query}
        onChange={onChange}
        onFocus={() => {
          if (results) setOpen(true);
        }}
        onKeyDown={onKeyDown}
        autoFocus={autoFocus}
        placeholder={t("header.searchPlaceholder")}
        aria-label={t("header.search")}
        role="combobox"
        aria-expanded={open}
        aria-controls={listId}
        aria-activedescendant={
          open && active >= 0 ? `${listId}-${active}` : undefined
        }
        aria-autocomplete="list"
        className={`field w-full pl-[33px] pr-[14px] ${
          large ? "py-[10px] text-[15px]" : "py-[9px] text-sm"
        }`}
      />

      {open && (
        <div className="absolute left-0 right-0 top-[calc(100%+8px)] z-30 overflow-hidden rounded-[14px] border border-border bg-surface shadow-[0_18px_50px_-18px_rgba(0,0,0,.35)] animate-fade-up">
          {shown.length === 0 ? (
            <div className="px-4 py-3.5 text-[13.5px] text-muted">
              {loading ? t("search.searching") : t("search.noResults", { term: q })}
            </div>
          ) : (
            <>
              <ul
                role="listbox"
                id={listId}
                aria-label={t("header.search")}
                className="max-h-[min(420px,60vh)] overflow-y-auto p-1.5"
              >
                {shown.map((a, i) => (
                  <li key={a.slug} role="option" aria-selected={i === active} id={`${listId}-${i}`}>
                    <button
                      type="button"
                      onClick={() => pick(a, i)}
                      onMouseMove={() => setActive(i)}
                      className={`block w-full rounded-[10px] px-3 py-2.5 text-left transition-colors ${
                        i === active ? "bg-hoverbg" : ""
                      }`}
                    >
                      <div className="truncate text-[14.5px] font-semibold text-text">
                        {a.title}
                      </div>
                      <div className="mt-0.5 line-clamp-2 text-[13px] leading-[1.5] text-muted">
                        {a.excerpt}
                      </div>
                      <div className="mt-1.5 flex items-center gap-2 text-[12px] text-faint">
                        <span className="chip-accent px-2 py-[2px] text-[11px]">
                          {a.category}
                        </span>
                        <span>{a.read}</span>
                      </div>
                    </button>
                  </li>
                ))}
              </ul>
              <button
                type="button"
                onClick={viewAll}
                className="block w-full border-t border-border px-4 py-2.5 text-center text-[13px] font-semibold text-accent-ink transition-colors hover:bg-hoverbg"
              >
                {t("search.viewAll", { count: results?.length ?? 0 })}
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}
