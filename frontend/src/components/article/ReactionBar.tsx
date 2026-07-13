"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";
import type { ReactionStatus } from "@/lib/types";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useT } from "@/lib/i18n/provider";
import { track } from "@/lib/analytics";

// Outline/filled pairs re-color with the theme via `currentColor`, matching
// ShareBar's inline-SVG approach (no icon dependency).
function HeartIcon({ filled }: { filled: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="17"
      height="17"
      fill={filled ? "currentColor" : "none"}
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z" />
    </svg>
  );
}

function BookmarkIcon({ filled }: { filled: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="17"
      height="17"
      fill={filled ? "currentColor" : "none"}
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M19 21l-7-4-7 4V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v16z" />
    </svg>
  );
}

/**
 * Like + bookmark toggles for an article. The like count is public; toggling
 * requires login — anonymous readers get an inline login prompt instead of a
 * surprise redirect. State is confirmed from the server response after each
 * toggle so the count can't drift.
 */
export function ReactionBar({
  slug,
  initial,
}: {
  slug: string;
  initial: ReactionStatus;
}) {
  const t = useT();
  const { user } = useAuth();
  const [status, setStatus] = useState<ReactionStatus>(initial);
  const [pending, setPending] = useState<"like" | "bookmark" | null>(null);
  const [showLoginHint, setShowLoginHint] = useState(false);
  const hintTimer = useRef<number | undefined>(undefined);

  // Clear the auto-hide timer on unmount so it can't fire after teardown.
  useEffect(() => () => window.clearTimeout(hintTimer.current), []);

  const toggle = useCallback(
    async (kind: "like" | "bookmark") => {
      if (!user) {
        setShowLoginHint(true);
        window.clearTimeout(hintTimer.current);
        hintTimer.current = window.setTimeout(
          () => setShowLoginHint(false),
          5000,
        );
        return;
      }
      if (pending) return;
      const on = kind === "like" ? !status.liked : !status.bookmarked;
      setPending(kind);
      try {
        const next = await api.setReaction(slug, kind, on);
        setStatus(next);
        track(kind === "like" ? "like_article" : "bookmark_article", {
          slug,
          on,
        });
      } catch {
        // Leave the previous state; the buttons stay usable for a retry.
      } finally {
        setPending(null);
      }
    },
    [user, pending, status.liked, status.bookmarked, slug],
  );

  const pill =
    "inline-flex items-center gap-2 rounded-full border px-3 py-2 text-[14px] font-semibold transition-colors sm:px-4";
  const off = "border-border text-c3a hover:border-accent-ink hover:text-accent-ink";
  const on = "border-accent-ink text-accent-ink";

  const likeLabel = status.liked
    ? t("article.reactions.unlike")
    : t("article.reactions.like");
  const saveLabel = status.bookmarked
    ? t("article.reactions.unsave")
    : t("article.reactions.save");

  return (
    <div className="relative flex flex-wrap items-center gap-x-3 gap-y-2.5">
      <button
        type="button"
        onClick={() => toggle("like")}
        disabled={pending === "like"}
        aria-pressed={status.liked}
        aria-label={likeLabel}
        title={likeLabel}
        className={`${pill} ${status.liked ? on : off}`}
      >
        <HeartIcon filled={status.liked} />
        {/* Text label collapses to icon-only on phones; title covers hover. */}
        <span className="hidden sm:inline">{t("article.reactions.like")}</span>
        <span
          className={`text-[13.5px] font-bold ${status.liked ? "" : "text-muted"}`}
        >
          {status.likes}
        </span>
      </button>

      <button
        type="button"
        onClick={() => toggle("bookmark")}
        disabled={pending === "bookmark"}
        aria-pressed={status.bookmarked}
        aria-label={saveLabel}
        title={saveLabel}
        className={`${pill} ${status.bookmarked ? on : off}`}
      >
        <BookmarkIcon filled={status.bookmarked} />
        <span className="hidden sm:inline">
          {status.bookmarked
            ? t("article.reactions.saved")
            : t("article.reactions.save")}
        </span>
      </button>

      {/* Floating popover (not in flow) so it never reflows the action bar. */}
      {showLoginHint && !user && (
        <span
          role="status"
          className="absolute left-0 top-[calc(100%+8px)] z-10 w-max max-w-[calc(100vw-48px)] rounded-[10px] border border-border bg-surface px-3 py-2 text-[13px] text-muted shadow-[0_12px_34px_-14px_rgba(0,0,0,.35)] animate-fade-up"
        >
          {t("article.reactions.loginPrompt")}{" "}
          <Link
            href="/login"
            className="font-semibold text-accent-ink no-underline hover:opacity-75"
          >
            {t("article.reactions.loginCta")}
          </Link>
        </span>
      )}
    </div>
  );
}
