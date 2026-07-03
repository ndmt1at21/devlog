"use client";

import Link from "next/link";
import type { ArticleSummary } from "@/lib/types";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

export function FeaturedCard({
  article,
  position = 0,
}: {
  article: ArticleSummary;
  position?: number;
}) {
  const t = useT();
  return (
    <Link
      href={`/articles/${article.slug}`}
      onClick={() =>
        track("select_article", {
          slug: article.slug,
          title: article.title,
          category: article.category,
          list: "featured",
          position,
        })
      }
      className="grid grid-cols-1 gap-0 overflow-hidden rounded-[18px] border border-border bg-surface p-3.5 no-underline transition-all hover:border-hover hover:shadow-[0_14px_40px_-22px_rgba(0,0,0,.22)] md:grid-cols-[1.15fr_1fr]"
    >
      <div className="flex flex-col justify-center p-[26px]">
        <span className="chip-accent mb-4 self-start px-3 py-[5px] text-[12.5px]">
          {article.category}
        </span>
        <h1 className="mb-3.5 text-[28px] font-extrabold leading-[1.18] tracking-[-.025em] text-balance text-text md:text-[32px]">
          {article.title}
        </h1>
        <p className="mb-[22px] text-[16.5px] leading-[1.6] text-pretty text-c5b">
          {article.excerpt}
        </p>
        <div className="flex flex-wrap items-center gap-2.5 text-[14px] text-muted">
          <span className="flex h-[30px] w-[30px] items-center justify-center rounded-full bg-chip text-[13px] font-bold text-strong">
            {article.authorInitial}
          </span>
          <span className="font-semibold text-c3a">{article.author}</span>
          <span>·</span>
          <span>{article.date}</span>
          <span>·</span>
          <span>{article.read}</span>
        </div>
      </div>
      <div className="cover-hatch relative flex min-h-[220px] items-center justify-center rounded-xl">
        <span className="font-mono text-[12px] tracking-[.04em] text-mono">
          {t("home.cover")}
        </span>
      </div>
    </Link>
  );
}
