import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { fetchArticle, fetchComments, fetchReactions } from "@/lib/server-api";
import { getLocale } from "@/lib/i18n/server";
import { highlightCode } from "@/lib/shiki";
import { ArticleView } from "@/components/article/ArticleView";
import { JsonLd } from "@/components/seo/JsonLd";
import { SITE_NAME, absoluteUrl, organization } from "@/lib/seo";
import type { ArticleDetail } from "@/lib/types";

type Params = { params: Promise<{ slug: string }> };

// resolveLang picks the title/excerpt/body to surface for a locale, falling back
// to the article's primary language when that locale has no translation.
function resolveLang(detail: ArticleDetail, locale: string) {
  const available = detail.availableLangs ?? [detail.lang];
  const lang = available.includes(locale) ? locale : detail.lang;
  const tr = lang !== detail.lang ? detail.translations?.[lang] : undefined;
  return {
    lang,
    title: tr?.title ?? detail.title,
    excerpt: tr?.excerpt ?? detail.excerpt,
  };
}

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const { slug } = await params;
  const article = await fetchArticle(slug);
  if (!article) return { title: "Không tìm thấy bài viết" };

  // Metadata follows the reader's UI locale when that translation exists.
  const v = resolveLang(article, await getLocale());
  const url = `/articles/${slug}`;
  // Prefer the article cover; fall back to the branded default OG image.
  const images = article.cover ? [article.cover] : undefined;

  return {
    title: v.title,
    description: v.excerpt,
    alternates: { canonical: url },
    openGraph: {
      type: "article",
      url,
      title: v.title,
      description: v.excerpt,
      authors: [article.author],
      tags: article.tags,
      publishedTime: article.publishedAt || undefined,
      images,
    },
    twitter: {
      card: "summary_large_image",
      title: v.title,
      description: v.excerpt,
      images,
    },
  };
}

export default async function ArticlePage({ params }: Params) {
  const { slug } = await params;
  const detail = await fetchArticle(slug);
  if (!detail) notFound();

  const locale = await getLocale();

  // Pre-highlight code blocks on the server so no highlighting JS ships. Every
  // language's body is highlighted, since the reader can toggle language
  // client-side without a refetch.
  const allBodies = [
    detail.body,
    ...Object.values(detail.translations ?? {}).map((t) => t.body),
  ];
  await Promise.all(
    allBodies.flat().map(async (block) => {
      if (block.type === "code" && block.code) {
        block.html = await highlightCode(block.code, block.lang);
      }
    }),
  );

  const [comments, reactions] = await Promise.all([
    detail.locked ? Promise.resolve([]) : fetchComments(slug),
    fetchReactions(slug),
  ]);

  const v = resolveLang(detail, locale);
  const articleUrl = absoluteUrl(`/articles/${slug}`);
  const blogPosting = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    mainEntityOfPage: { "@type": "WebPage", "@id": articleUrl },
    url: articleUrl,
    headline: v.title,
    description: v.excerpt,
    ...(detail.cover ? { image: [detail.cover] } : {}),
    ...(detail.publishedAt ? { datePublished: detail.publishedAt } : {}),
    author: { "@type": "Person", name: detail.author },
    publisher: organization,
    keywords: detail.tags.join(", "),
    articleSection: detail.category,
    inLanguage: v.lang,
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
        name: v.title,
        item: articleUrl,
      },
    ],
  };

  return (
    <>
      <JsonLd data={blogPosting} />
      <JsonLd data={breadcrumbs} />
      <ArticleView
        detail={detail}
        initialComments={comments}
        initialReactions={reactions}
        initialLang={locale}
      />
    </>
  );
}
