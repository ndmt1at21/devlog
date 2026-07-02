"use client";

import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { absoluteUrl } from "@/lib/seo";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

// Brand + UI glyphs as inline SVG so the bar ships no icon dependency and
// re-colors with the theme via `currentColor`.
function XIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">
      <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
    </svg>
  );
}

function FacebookIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">
      <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" />
    </svg>
  );
}

function LinkedInIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">
      <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.225 0z" />
    </svg>
  );
}

function LinkIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      width="18"
      height="18"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      width="18"
      height="18"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}

// Web Share API support is a static browser capability, so a no-op subscription
// is enough — useSyncExternalStore reads `false` on the server and the real
// value on the client without a hydration mismatch.
const subscribeNoop = () => () => {};
const getNativeShareSupport = () =>
  typeof navigator !== "undefined" && typeof navigator.share === "function";

function ShareIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      width="16"
      height="16"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8" />
      <path d="M16 6l-4-4-4 4" />
      <path d="M12 2v13" />
    </svg>
  );
}

/**
 * Social-share row for an article. Mobile browsers get the native share sheet
 * via the Web Share API; everywhere else (and as an always-available fallback)
 * there are explicit X / Facebook / LinkedIn intents plus copy-link.
 */
export function ShareBar({
  slug,
  title,
  excerpt,
}: {
  slug: string;
  title: string;
  excerpt: string;
}) {
  const t = useT();
  // Encode the slug segment so odd slugs (`#`, `?`, spaces) can't truncate the
  // shared/copied link — matching how api.ts / server-api.ts build slug paths.
  const url = absoluteUrl(`/articles/${encodeURIComponent(slug)}`);
  const canNativeShare = useSyncExternalStore(
    subscribeNoop,
    getNativeShareSupport,
    () => false,
  );
  const [copied, setCopied] = useState(false);
  // Bumped on every successful copy so the live region re-announces even when
  // the "Copied" text is unchanged (e.g. copying twice within the reset window).
  const [announceKey, setAnnounceKey] = useState(0);
  const resetTimer = useRef<number | undefined>(undefined);

  // Clear any pending reset on unmount so the timer can't fire after teardown.
  useEffect(() => () => window.clearTimeout(resetTimer.current), []);

  const nativeShare = useCallback(async () => {
    try {
      await navigator.share({ title, text: excerpt, url });
      track("share_article", { slug, method: "native" });
    } catch {
      // Swallow: the user likely dismissed the sheet (AbortError) or the
      // platform rejected the payload. The explicit buttons remain available.
    }
  }, [title, excerpt, url, slug]);

  const copyLink = useCallback(async () => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(url);
      } else {
        // Fallback for non-secure contexts / older browsers.
        const ta = document.createElement("textarea");
        ta.value = url;
        ta.setAttribute("readonly", "");
        ta.style.position = "fixed";
        ta.style.left = "-9999px";
        document.body.appendChild(ta);
        ta.select();
        const ok = document.execCommand("copy");
        document.body.removeChild(ta);
        // execCommand returns false (rather than throwing) when the copy is
        // rejected — treat that as failure so we don't show a false confirmation.
        if (!ok) throw new Error("copy command was rejected");
      }
      setCopied(true);
      setAnnounceKey((k) => k + 1);
      track("share_article", { slug, method: "copy" });
      window.clearTimeout(resetTimer.current);
      resetTimer.current = window.setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard blocked — nothing else we can do gracefully.
    }
  }, [url, slug]);

  const targets = [
    {
      key: "twitter" as const,
      label: t("article.share.twitter"),
      href: `https://twitter.com/intent/tweet?text=${encodeURIComponent(title)}&url=${encodeURIComponent(url)}`,
      icon: <XIcon />,
    },
    {
      key: "facebook" as const,
      label: t("article.share.facebook"),
      href: `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}`,
      icon: <FacebookIcon />,
    },
    {
      key: "linkedin" as const,
      label: t("article.share.linkedin"),
      href: `https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(url)}`,
      icon: <LinkedInIcon />,
    },
  ];

  const iconBtn =
    "inline-flex h-10 w-10 items-center justify-center rounded-full border border-border text-c3a transition-colors hover:border-accent-ink hover:text-accent-ink";

  return (
    <div className="flex flex-wrap items-center gap-2.5">
        {canNativeShare && (
          <button
            type="button"
            onClick={nativeShare}
            className="btn-accent inline-flex items-center gap-2 px-4 py-2 text-[14px]"
          >
            <ShareIcon />
            {t("article.share.native")}
          </button>
        )}

        {targets.map((target) => (
          <a
            key={target.key}
            href={target.href}
            target="_blank"
            rel="noopener noreferrer"
            aria-label={target.label}
            title={target.label}
            onClick={() => track("share_article", { slug, method: target.key })}
            className={iconBtn}
          >
            {target.icon}
          </a>
        ))}

        <button
          type="button"
          onClick={copyLink}
          aria-label={t("article.share.copy")}
          title={t("article.share.copy")}
          className={`${iconBtn} ${copied ? "border-accent-ink text-accent-ink" : ""}`}
        >
          {copied ? <CheckIcon /> : <LinkIcon />}
        </button>

        {/* Single source of truth for the copy confirmation: a persistent live
            region (so it's registered before it changes) whose keyed child
            swaps on each copy — the button's own label stays static to avoid a
            duplicate focused-element announcement. */}
        <span role="status" aria-live="polite" aria-atomic="true" className="sr-only">
          {copied ? (
            <span key={announceKey}>{t("article.share.copied")}</span>
          ) : null}
        </span>
    </div>
  );
}
