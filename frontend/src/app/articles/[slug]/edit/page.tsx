import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { fetchArticle } from "@/lib/server-api";
import { NewArticleForm } from "@/components/editor/NewArticleForm";

type Params = { params: Promise<{ slug: string }> };

// Private authoring surface — keep it out of search indexes and the sitemap.
export const metadata: Metadata = {
  title: "Chỉnh sửa bài viết",
  robots: { index: false, follow: false },
};

export default async function EditArticlePage({ params }: Params) {
  const { slug } = await params;
  const article = await fetchArticle(slug);
  if (!article) notFound();

  // Login / permission / authorship are enforced inside the form (and the
  // backend re-checks ownership on PUT), so the article body is safe to prefill.
  return <NewArticleForm article={article} />;
}
