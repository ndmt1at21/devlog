import {
  fetchArticles,
  fetchCategories,
  fetchFeatured,
} from "@/lib/server-api";
import { HomeContent } from "@/components/home/HomeContent";

export default async function Home({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  const { q } = await searchParams;
  const query = q?.trim() ?? "";

  const [featured, articles, categories] = await Promise.all([
    fetchFeatured(),
    fetchArticles(query ? { q: query } : undefined),
    fetchCategories(),
  ]);

  return (
    <HomeContent
      featured={featured}
      articles={articles}
      categories={categories}
      query={query}
    />
  );
}
