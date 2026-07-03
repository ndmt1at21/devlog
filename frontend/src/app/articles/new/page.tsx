import type { Metadata } from "next";
import { NewArticleForm } from "@/components/editor/NewArticleForm";

// Private authoring surface — keep it out of search indexes and the sitemap.
export const metadata: Metadata = {
  title: "Viết bài mới",
  robots: { index: false, follow: false },
};

export default function NewArticlePage() {
  return <NewArticleForm />;
}
