"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import type { ArticleSummary } from "@/lib/types";
import { useSearch } from "@/lib/search";
import { useT } from "@/lib/i18n/provider";
import { track } from "@/lib/analytics";
import { FeaturedCard } from "./FeaturedCard";
import { ArticleCard } from "./ArticleCard";
import { CategoryFilter } from "./CategoryFilter";

const ALL = "Tất cả";

function matches(a: ArticleSummary, q: string): boolean {
  const hay = `${a.title} ${a.excerpt} ${a.category} ${a.tags.join(" ")}`.toLowerCase();
  return hay.includes(q);
}

export function HomeContent({
  featured,
  articles,
  categories,
}: {
  featured: ArticleSummary | null;
  articles: ArticleSummary[];
  categories: string[];
}) {
  const { query } = useSearch();
  const t = useT();
  const [category, setCategory] = useState(ALL);

  const q = query.trim().toLowerCase();
  const searching = q !== "";
  const showFeatured = !searching && category === ALL && !!featured;

  const grid = useMemo(() => {
    let list = articles;
    if (category !== ALL) list = list.filter((a) => a.category === category);
    if (q) list = list.filter((a) => matches(a, q));
    if (showFeatured && featured)
      list = list.filter((a) => a.slug !== featured.slug);
    return list;
  }, [articles, category, q, showFeatured, featured]);

  // Debounced `search` analytics event.
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    if (!searching) return;
    if (debounce.current) clearTimeout(debounce.current);
    debounce.current = setTimeout(() => {
      track("search", { search_term: q, results_count: grid.length });
    }, 500);
    return () => {
      if (debounce.current) clearTimeout(debounce.current);
    };
  }, [q, searching, grid.length]);

  return (
    <div>
      {showFeatured && featured && (
        <section className="mx-auto max-w-[1120px] px-6 pt-11">
          <div className="mb-5 text-[13px] font-semibold uppercase tracking-[.08em] text-accent-ink">
            {t("home.featured")}
          </div>
          <FeaturedCard article={featured} />
        </section>
      )}

      <CategoryFilter
        categories={categories}
        active={category}
        onSelect={setCategory}
      />

      {searching && (
        <div className="mx-auto mt-10 max-w-[1120px] px-6">
          <div className="text-[15px] font-medium text-c5b">
            {grid.length > 0
              ? t("home.resultsFor", { count: grid.length, term: query.trim() })
              : t("home.noResultsFor", { term: query.trim() })}
          </div>
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
