"use client";

import Link from "next/link";
import type { PartLink } from "@/lib/types";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

export function SeriesNav({
  fromSlug,
  prev,
  next,
}: {
  fromSlug: string;
  prev?: PartLink;
  next?: PartLink;
}) {
  const t = useT();
  if (!prev && !next) return null;

  const nav = (direction: "prev" | "next", to: string) =>
    track("series_nav", { direction, from_slug: fromSlug, to_slug: to });

  return (
    <div className="mt-[42px] flex flex-wrap gap-3">
      {prev && (
        <Link
          href={`/articles/${prev.id}`}
          onClick={() => nav("prev", prev.id)}
          className="min-w-[200px] flex-1 rounded-xl border border-border2 bg-surface px-4 py-3.5 text-left no-underline transition-colors hover:border-hover"
        >
          <span className="mb-[3px] block text-[12px] text-faint">
            {t("article.prevPart")}
          </span>
          <span className="block text-[14.5px] font-semibold text-text">
            {prev.ptitle}
          </span>
        </Link>
      )}
      {next && (
        <Link
          href={`/articles/${next.id}`}
          onClick={() => nav("next", next.id)}
          className="min-w-[200px] flex-1 rounded-xl border-none bg-accent px-4 py-3.5 text-right no-underline transition-colors hover:brightness-95"
        >
          <span className="mb-[3px] block text-[12px] text-[rgba(44,35,0,.72)]">
            {t("article.nextPart")}
          </span>
          <span className="block text-[14.5px] font-semibold text-on-accent">
            {next.ptitle}
          </span>
        </Link>
      )}
    </div>
  );
}
