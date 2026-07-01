import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { fetchArticle, fetchComments } from "@/lib/server-api";
import { highlightCode } from "@/lib/highlight";
import { ArticleView } from "@/components/article/ArticleView";
import { JsonLd } from "@/components/seo/JsonLd";
import { SITE_NAME, absoluteUrl, organization } from "@/lib/seo";

type Params = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const { slug } = await params;
  const article = await fetchArticle(slug);
  if (!article) return { title: "Không tìm thấy bài viết" };

  const url = `/articles/${slug}`;
  // Prefer the article cover; fall back to the branded default OG image.
  const images = article.cover ? [article.cover] : undefined;

  return {
    title: article.title,
    description: article.excerpt,
    alternates: { canonical: url },
    openGraph: {
      type: "article",
      url,
      title: article.title,
      description: article.excerpt,
      authors: [article.author],
      tags: article.tags,
      publishedTime: article.publishedAt || undefined,
      images,
    },
    twitter: {
      card: "summary_large_image",
      title: article.title,
      description: article.excerpt,
      images,
    },
  };
}

export default async function ArticlePage({ params }: Params) {
  const { slug } = await params;
  const detail = await fetchArticle(slug);
  if (!detail) notFound();

  // Pre-highlight code blocks on the server so no highlighting JS ships.
  await Promise.all(
    detail.body.map(async (block) => {
      if (block.type === "code" && block.code) {
        block.html = await highlightCode(block.code, block.lang);
      }
    }),
  );

  const comments = detail.locked ? [] : await fetchComments(slug);

  const articleUrl = absoluteUrl(`/articles/${slug}`);
  const blogPosting = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    mainEntityOfPage: { "@type": "WebPage", "@id": articleUrl },
    url: articleUrl,
    headline: detail.title,
    description: detail.excerpt,
    ...(detail.cover ? { image: [detail.cover] } : {}),
    ...(detail.publishedAt ? { datePublished: detail.publishedAt } : {}),
    author: { "@type": "Person", name: detail.author },
    publisher: organization,
    keywords: detail.tags.join(", "),
    articleSection: detail.category,
    inLanguage: "vi",
    // Locked articles ship only a truncated preview (the gated body is dropped
    // server-side), so we honestly flag them as not free. We omit the paywalled
    // `hasPart`/cssSelector form because that requires the full body to be
    // present-but-hidden in the DOM, which it isn't here.
    isAccessibleForFree: !detail.locked,
  };
  const breadcrumbs = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: [
      { "@type": "ListItem", position: 1, name: SITE_NAME, item: absoluteUrl("/") },
      {
        "@type": "ListItem",
        position: 2,
        name: detail.title,
        item: articleUrl,
      },
    ],
  };

  return (
    <>
      <JsonLd data={blogPosting} />
      <JsonLd data={breadcrumbs} />
      <ArticleView detail={detail} initialComments={comments} />
    </>
  );
}
