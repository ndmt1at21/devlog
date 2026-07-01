"use client";

import Link from "next/link";
import { useEffect } from "react";
import type { ArticleDetail, Block, Comment } from "@/lib/types";
import { useAuth } from "@/lib/auth";
import { useT } from "@/lib/i18n/provider";
import { track } from "@/lib/analytics";
import { CodeBlock } from "./blocks/CodeBlock";
import { Diagram } from "./blocks/Diagram";
import { AdSlot } from "./blocks/AdSlot";
import { TagList } from "./TagList";
import { SeriesBox } from "./SeriesBox";
import { Paywall } from "./Paywall";
import { SeriesNav } from "./SeriesNav";
import { ShareBar } from "./ShareBar";
import { Comments } from "./Comments";
import { ScrollDepthTracker } from "@/components/analytics/ScrollDepthTracker";

// Ads appear after the 4th block (index 3) for non-premium readers, mirroring
// the mockup's `showAds && !premium && blocks>3` rule.
const AD_INDEX = 3;

function renderBlock(block: Block, i: number, slug: string) {
  switch (block.type) {
    case "h":
      return (
        <h2
          key={i}
          className="mb-2.5 mt-11 text-[25px] font-bold leading-[1.3] tracking-[-.02em] text-text"
        >
          {block.text}
        </h2>
      );
    case "quote":
      return (
        <blockquote
          key={i}
          className="my-[30px] border-l-[3px] border-accent py-1.5 pl-[22px] text-[20px] font-medium leading-[1.6] text-c3a"
        >
          {block.text}
        </blockquote>
      );
    case "code":
      return (
        <CodeBlock
          key={i}
          lang={block.lang}
          code={block.code ?? ""}
          html={block.html}
          slug={slug}
        />
      );
    case "diagram":
      return (
        <Diagram key={i} steps={block.steps ?? []} caption={block.caption} />
      );
    case "p":
    default:
      return (
        <p key={i} className="mb-[22px]">
          {block.text}
        </p>
      );
  }
}

export function ArticleView({
  detail,
  initialComments,
}: {
  detail: ArticleDetail;
  initialComments: Comment[];
}) {
  const { premium } = useAuth();
  const t = useT();

  useEffect(() => {
    track("view_article", {
      slug: detail.slug,
      category: detail.category,
      series_slug: detail.series,
      is_premium: premium,
    });
  }, [detail.slug, detail.category, detail.series, premium]);

  const showAds = !premium && !detail.locked && detail.body.length > AD_INDEX;

  return (
    <article className="mx-auto max-w-[704px] px-6 pb-24 pt-9">
      <Link
        href="/"
        className="mb-[26px] inline-flex items-center gap-1.5 text-[14.5px] font-semibold text-accent-ink no-underline transition-opacity hover:opacity-75"
      >
        {t("common.back")}
      </Link>

      <h1 className="mb-4 text-[32px] font-extrabold leading-[1.16] tracking-[-.03em] text-balance text-text md:text-[38px]">
        {detail.title}
      </h1>

      <TagList tags={detail.tags} />

      <div className="flex flex-wrap items-center gap-[11px] border-b border-border pb-[30px] text-[14.5px] text-muted">
        <span className="avatar-accent h-9 w-9 text-[14px]">
          {detail.authorInitial}
        </span>
        <span className="font-semibold text-c3a">{detail.author}</span>
        <span>·</span>
        <span>{detail.date}</span>
        <span>·</span>
        <span>{detail.read}</span>
      </div>

      {detail.inSeries && detail.seriesParts && (
        <SeriesBox
          title={detail.seriesTitle ?? ""}
          partLabel={detail.seriesPartLabel ?? ""}
          parts={detail.seriesParts}
        />
      )}

      <div className="cover-hatch relative my-[30px] mb-3.5 flex h-[300px] items-center justify-center rounded-[14px]">
        <span className="font-mono text-[12px] tracking-[.04em] text-mono">
          {t("article.cover")}
        </span>
      </div>

      <div className="text-[19px] leading-[1.85] text-body">
        {detail.body.map((block, i) => (
          <div key={i}>
            {renderBlock(block, i, detail.slug)}
            {showAds && i === AD_INDEX && <AdSlot slot="in-content" />}
          </div>
        ))}
      </div>

      {detail.locked && (
        <Paywall slug={detail.slug} seriesSlug={detail.series} />
      )}

      <ShareBar
        slug={detail.slug}
        title={detail.title}
        excerpt={detail.excerpt}
      />

      {detail.inSeries && (
        <SeriesNav
          fromSlug={detail.slug}
          prev={detail.prevPart}
          next={detail.nextPart}
        />
      )}

      {!detail.locked && (
        <Comments slug={detail.slug} initialComments={initialComments} />
      )}

      <ScrollDepthTracker slug={detail.slug} />
    </article>
  );
}
