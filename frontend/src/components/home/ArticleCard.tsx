"use client";

import Link from "next/link";
import type { ArticleSummary } from "@/lib/types";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

export function ArticleCard({
  article,
  position,
  list = "grid",
}: {
  article: ArticleSummary;
  position: number;
  list?: string;
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
          list,
          position,
        })
      }
      className="group flex flex-col rounded-2xl border border-border bg-surface p-[13px] text-left no-underline transition-all hover:-translate-y-0.5 hover:border-hover hover:shadow-[0_12px_34px_-22px_rgba(0,0,0,.2)]"
    >
      {article.cover ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={article.cover}
          alt=""
          className="mb-4 h-[172px] w-full rounded-[10px] border border-border object-cover"
        />
      ) : (
        <div className="cover-hatch relative mb-4 flex h-[172px] items-center justify-center rounded-[10px]">
          <span className="font-mono text-[11.5px] tracking-[.04em] text-mono">
            {t("home.cover")}
          </span>
        </div>
      )}
      <div className="flex flex-1 flex-col px-1.5 pb-2">
        <div className="mb-3 flex flex-wrap gap-1.5">
          <span className="chip-accent px-[11px] py-1 text-[12px]">
            {article.category}
          </span>
          {article.isSeries && article.seriesBadge && (
            <span className="chip-series px-[11px] py-1 text-[12px]">
              {article.seriesBadge}
            </span>
          )}
        </div>
        <h3 className="mb-[9px] text-[19px] font-bold leading-[1.28] tracking-[-.015em] text-balance text-text">
          {article.title}
        </h3>
        <p className="mb-4 flex-1 text-[14.5px] leading-[1.55] text-pretty text-muted">
          {article.excerpt}
        </p>
        <div className="flex items-center gap-[7px] text-[13px] text-subtle">
          <span className="font-semibold text-c5b">{article.author}</span>
          <span>·</span>
          <span>{article.read}</span>
        </div>
      </div>
    </Link>
  );
}
