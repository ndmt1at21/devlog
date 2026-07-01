import {
  fetchArticles,
  fetchCategories,
  fetchFeatured,
} from "@/lib/server-api";
import { HomeContent } from "@/components/home/HomeContent";

export default async function Home() {
  const [featured, articles, categories] = await Promise.all([
    fetchFeatured(),
    fetchArticles(),
    fetchCategories(),
  ]);

  return (
    <HomeContent
      featured={featured}
      articles={articles}
      categories={categories}
    />
  );
}
