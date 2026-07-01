import type { Metadata } from "next";
import { Be_Vietnam_Pro } from "next/font/google";
import { GoogleAnalytics } from "@next/third-parties/google";
import "./globals.css";
import { GoogleAdManager } from "@/components/ads/GoogleAdManager";
import { ThemeProvider, themeInitScript } from "@/components/theme/ThemeProvider";
import { AuthProvider } from "@/lib/auth";
import { SearchProvider } from "@/lib/search";
import { CoffeeProvider } from "@/components/coffee/CoffeeModal";
import { Header } from "@/components/layout/Header";
import { Footer } from "@/components/layout/Footer";
import { ScrollProgress } from "@/components/layout/ScrollProgress";
import { GA_ID } from "@/lib/analytics";
import { fetchMe } from "@/lib/server-api";
import { LocaleProvider } from "@/lib/i18n/provider";
import { getLocale, getT } from "@/lib/i18n/server";
import { JsonLd } from "@/components/seo/JsonLd";
import { SITE_NAME, SITE_URL, absoluteUrl, organization } from "@/lib/seo";

const beVietnam = Be_Vietnam_Pro({
  subsets: ["latin", "vietnamese"],
  weight: ["300", "400", "500", "600", "700", "800"],
  variable: "--font-be-vietnam",
  display: "swap",
});

export async function generateMetadata(): Promise<Metadata> {
  const [locale, t] = [await getLocale(), await getT()];
  return {
    metadataBase: new URL(
      process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000",
    ),
    title: {
      default: t("meta.titleDefault"),
      template: "%s · devnote",
    },
    description: t("meta.description"),
    alternates: { canonical: "/" },
    openGraph: {
      type: "website",
      siteName: "devnote",
      locale: locale === "en" ? "en_US" : "vi_VN",
    },
    twitter: {
      card: "summary_large_image",
    },
    robots: {
      index: true,
      follow: true,
    },
  };
}

export default async function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const [me, locale] = [await fetchMe(), await getLocale()];

  return (
    <html lang={locale} suppressHydrationWarning className={beVietnam.variable}>
      <body>
        <JsonLd
          data={{
            "@context": "https://schema.org",
            "@graph": [
              {
                "@type": "WebSite",
                name: SITE_NAME,
                url: SITE_URL,
                inLanguage: locale === "en" ? "en" : "vi",
                potentialAction: {
                  "@type": "SearchAction",
                  target: {
                    "@type": "EntryPoint",
                    urlTemplate: absoluteUrl("/?q={search_term_string}"),
                  },
                  "query-input": "required name=search_term_string",
                },
              },
              organization,
            ],
          }}
        />
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
        <LocaleProvider initialLocale={locale}>
          <ThemeProvider>
            <AuthProvider initial={me}>
              <SearchProvider>
                <CoffeeProvider>
                  <Header />
                  <main>{children}</main>
                  <Footer />
                  <ScrollProgress />
                </CoffeeProvider>
              </SearchProvider>
            </AuthProvider>
          </ThemeProvider>
        </LocaleProvider>
        {GA_ID ? <GoogleAnalytics gaId={GA_ID} /> : null}
        <GoogleAdManager />
      </body>
    </html>
  );
}
