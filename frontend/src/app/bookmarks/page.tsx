import type { Metadata } from "next";
import Link from "next/link";
import { fetchBookmarks, fetchMe } from "@/lib/server-api";
import { ArticleCard } from "@/components/home/ArticleCard";
import { getT } from "@/lib/i18n/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getT();
  // Personal page: keep it out of search indexes.
  return { title: t("bookmarks.title"), robots: { index: false, follow: false } };
}

export default async function BookmarksPage() {
  const t = await getT();
  const [me, bookmarks] = await Promise.all([fetchMe(), fetchBookmarks()]);

  return (
    <section className="mx-auto max-w-[1120px] px-6 pb-20 pt-11">
      <h1 className="mb-2 text-[28px] font-extrabold tracking-[-.02em] text-text">
        {t("bookmarks.title")}
      </h1>
      <p className="mb-9 text-[15px] text-muted">{t("bookmarks.subtitle")}</p>

      {!me.authenticated ? (
        <div className="rounded-2xl border border-border bg-surface px-6 py-14 text-center">
          <h2 className="mb-2 text-[19px] font-bold text-text">
            {t("bookmarks.loginTitle")}
          </h2>
          <p className="mb-6 text-[14.5px] text-muted">
            {t("bookmarks.loginBody")}
          </p>
          <Link
            href="/login"
            className="btn-accent inline-flex px-5 py-2.5 text-[14.5px] no-underline"
          >
            {t("bookmarks.login")}
          </Link>
        </div>
      ) : !bookmarks || bookmarks.length === 0 ? (
        <div className="rounded-2xl border border-border bg-surface px-6 py-14 text-center">
          <p className="mb-5 text-[15px] text-muted">{t("bookmarks.empty")}</p>
          <Link
            href="/"
            className="text-[14.5px] font-semibold text-accent-ink no-underline hover:opacity-75"
          >
            {t("bookmarks.browse")}
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-[26px] sm:grid-cols-2 lg:grid-cols-3">
          {bookmarks.map((article, i) => (
            <ArticleCard
              key={article.slug}
              article={article}
              position={i + 1}
              list="bookmarks"
            />
          ))}
        </div>
      )}
    </section>
  );
}
