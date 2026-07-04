"use client";

import { useMemo } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { ArticleSummary } from "@/lib/types";
import { useT } from "@/lib/i18n/provider";
import { FeaturedCarousel } from "./FeaturedCarousel";
import { ArticleCard } from "./ArticleCard";
import { CategoryFilter } from "./CategoryFilter";

const ALL = "Tất cả";

/**
 * Homepage below the header. Both the text search (?q=…) and the category
 * filter (?category=…) live in the URL, so a reload — or back/forward —
 * restores the current view. Search is applied server-side (`articles` arrive
 * pre-filtered); the category refines that list client-side.
 */
export function HomeContent({
  featured,
  articles,
  categories,
  query = "",
  category = ALL,
}: {
  featured: ArticleSummary[];
  articles: ArticleSummary[];
  categories: string[];
  query?: string;
  category?: string;
}) {
  const t = useT();
  const router = useRouter();

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

  // Write the selection to the URL (preserving the search term) instead of
  // holding it in local state, so it survives a reload. The article data does
  // not depend on ?category=, so this is a cheap soft navigation; keep the
  // scroll position so switching filters doesn't jump to the top.
  const selectCategory = (cat: string) => {
    const params = new URLSearchParams();
    if (query) params.set("q", query);
    if (cat !== ALL) params.set("category", cat);
    const qs = params.toString();
    router.push(qs ? `/?${qs}` : "/", { scroll: false });
  };

  // Clearing the search drops ?q= but keeps the active category filter.
  const clearSearchHref =
    category !== ALL ? `/?category=${encodeURIComponent(category)}` : "/";

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
        onSelect={selectCategory}
      />

      {searching && (
        <div className="mx-auto mt-10 flex max-w-[1120px] flex-wrap items-center gap-x-4 gap-y-2 px-6">
          <div className="text-[15px] font-medium text-c5b">
            {grid.length > 0
              ? t("home.resultsFor", { count: grid.length, term: query })
              : t("home.noResultsFor", { term: query })}
          </div>
          <Link
            href={clearSearchHref}
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
