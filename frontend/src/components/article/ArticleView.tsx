"use client";

import Link from "next/link";
import { useEffect } from "react";
import type { ArticleDetail, Comment, ReactionStatus } from "@/lib/types";
import { useAuth } from "@/lib/auth";
import { useT } from "@/lib/i18n/provider";
import { track } from "@/lib/analytics";
import { BlockView } from "./BlockRenderer";
import { AdSlot } from "./blocks/AdSlot";
import { TagList } from "./TagList";
import { SeriesBox } from "./SeriesBox";
import { Paywall } from "./Paywall";
import { SeriesNav } from "./SeriesNav";
import { ShareBar } from "./ShareBar";
import { ReactionBar } from "./ReactionBar";
import { Comments } from "./Comments";
import { ScrollDepthTracker } from "@/components/analytics/ScrollDepthTracker";

// Ads appear after the 4th block (index 3) for non-premium readers, mirroring
// the mockup's `showAds && !premium && blocks>3` rule.
const AD_INDEX = 3;

export function ArticleView({
  detail,
  initialComments,
  initialReactions,
}: {
  detail: ArticleDetail;
  initialComments: Comment[];
  initialReactions: ReactionStatus;
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

      {/* break-words (inherited) keeps long URLs/identifiers from overflowing
          the measure on small screens; pre/code are unaffected (nowrap). */}
      <div className="break-words text-[19px] leading-[1.85] text-body">
        {detail.body.map((block, i) => (
          <div key={i}>
            <BlockView block={block} slug={detail.slug} />
            {showAds && i === AD_INDEX && <AdSlot slot="in-content" />}
          </div>
        ))}
      </div>

      {detail.locked && (
        <Paywall slug={detail.slug} seriesSlug={detail.series} />
      )}

      {/* Action bar: like/save on the left, share cluster on the right. */}
      <div className="mt-12 flex flex-wrap items-center justify-between gap-x-6 gap-y-4 border-t border-border pt-7">
        <ReactionBar slug={detail.slug} initial={initialReactions} />
        <ShareBar
          slug={detail.slug}
          title={detail.title}
          excerpt={detail.excerpt}
        />
      </div>

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
