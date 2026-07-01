import type { MetadataRoute } from "next";
import { SITE_URL } from "@/lib/seo";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      // Non-content routes: the API, the payment-return landing, and auth
      // flows. Keeping these out of the index avoids thin/duplicate pages.
      disallow: [
        "/api/",
        "/coffee/result",
        "/login",
        "/register",
        "/forgot-password",
      ],
    },
    sitemap: `${SITE_URL}/sitemap.xml`,
    host: SITE_URL,
  };
}
