"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import type { ArticleSummary } from "@/lib/types";
import { useT } from "@/lib/i18n/provider";
import { FeaturedCarousel } from "./FeaturedCarousel";
import { ArticleCard } from "./ArticleCard";
import { CategoryFilter } from "./CategoryFilter";

const ALL = "Tất cả";

/**
 * Homepage below the header. Text search is server-side: the page is fetched
 * with /?q=… (from the header SearchBox or a tag click) and `articles` arrive
 * pre-filtered. The category filter refines client-side on top of that.
 */
export function HomeContent({
  featured,
  articles,
  categories,
  query = "",
}: {
  featured: ArticleSummary[];
  articles: ArticleSummary[];
  categories: string[];
  query?: string;
}) {
  const t = useT();
  const [category, setCategory] = useState(ALL);

  const searching = query !== "";
  // Featured stays pinned on top across category switches; only an active
  // search takes over the page.
  const showFeatured = !searching && featured.length > 0;

  const grid = useMemo(() => {
    let list = articles;
    if (category !== ALL) list = list.filter((a) => a.category === category);
    if (showFeatured) {
      const pinned = new Set(featured.map((f) => f.slug));
      list = list.filter((a) => !pinned.has(a.slug));
    }
    return list;
  }, [articles, category, showFeatured, featured]);

  return (
    <div>
      {showFeatured && (
        <section className="mx-auto max-w-[1120px] px-6 pt-11">
          <div className="mb-5 text-[13px] font-semibold uppercase tracking-[.08em] text-accent-ink">
            {t("home.featured")}
          </div>
          <FeaturedCarousel articles={featured} />
        </section>
      )}

      <CategoryFilter
        categories={categories}
        active={category}
        onSelect={setCategory}
      />

      {searching && (
        <div className="mx-auto mt-10 flex max-w-[1120px] flex-wrap items-center gap-x-4 gap-y-2 px-6">
          <div className="text-[15px] font-medium text-c5b">
            {grid.length > 0
              ? t("home.resultsFor", { count: grid.length, term: query })
              : t("home.noResultsFor", { term: query })}
          </div>
          <Link
            href="/"
            className="text-[13.5px] font-semibold text-accent-ink no-underline hover:opacity-75"
          >
            {t("home.clearSearch")}
          </Link>
        </div>
      )}

      <section className="mx-auto mt-7 max-w-[1120px] px-6 pb-20">
        <div className="grid grid-cols-1 gap-[26px] sm:grid-cols-2 lg:grid-cols-3">
          {grid.map((a, i) => (
            <ArticleCard
              key={a.slug}
              article={a}
              position={i}
              list={searching ? "search" : "grid"}
            />
          ))}
        </div>
        {grid.length === 0 && (
          <div className="py-[60px] text-center text-[15px] text-subtle">
            {searching ? t("home.emptySearch") : t("home.emptyCategory")}
          </div>
        )}
      </section>
    </div>
  );
}
