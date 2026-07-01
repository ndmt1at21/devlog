"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { track } from "@/lib/analytics";
import { proEnabled } from "@/lib/features";
import { useT } from "@/lib/i18n/provider";

export function Paywall({
  slug,
  seriesSlug,
}: {
  slug: string;
  seriesSlug?: string;
}) {
  const router = useRouter();
  const t = useT();

  useEffect(() => {
    track("paywall_view", { slug, series_slug: seriesSlug });
  }, [slug, seriesSlug]);

  return (
    <div
      className="mt-1.5 rounded-[18px] px-7 py-[34px] text-center"
      style={{
        border: "1px solid color-mix(in srgb, var(--accent) 26%, transparent)",
        background: "color-mix(in srgb, var(--accent) 7%, var(--surface))",
      }}
    >
      <div aria-hidden="true" className="mb-2.5 text-[30px]">🔒</div>
      <h3 className="m-0 mb-2 text-[21px] font-extrabold tracking-[-.02em] text-text">
        {t("paywall.title")}
      </h3>
      <p className="mx-auto mb-5 max-w-[380px] text-[15px] leading-[1.6] text-c5b">
        {t("paywall.body")}
      </p>
      {proEnabled && (
        <button
          onClick={() => {
            track("paywall_upgrade_click", { slug, series_slug: seriesSlug });
            router.push("/pro");
          }}
          className="btn-accent px-[26px] py-3 text-[15px]"
        >
          {t("paywall.cta")}
        </button>
      )}
    </div>
  );
}
