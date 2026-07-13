import {
  fetchArticles,
  fetchCategories,
  fetchFeatured,
} from "@/lib/server-api";
import { HomeContent } from "@/components/home/HomeContent";

const ALL = "Tất cả";

export default async function Home({
  searchParams,
}: {
  searchParams: Promise<{ q?: string; category?: string }>;
}) {
  const { q, category } = await searchParams;
  const query = q?.trim() ?? "";

  const [featured, articles, categories] = await Promise.all([
    fetchFeatured(),
    fetchArticles(query ? { q: query } : undefined),
    fetchCategories(),
  ]);

  // The active category lives in the URL (?category=…) so a reload restores the
  // filter. Ignore an unknown/stale value rather than rendering an empty grid.
  const activeCategory =
    category && categories.includes(category) ? category : ALL;

  return (
    <HomeContent
      featured={featured}
      articles={articles}
      categories={categories}
      query={query}
      category={activeCategory}
    />
  );
}
