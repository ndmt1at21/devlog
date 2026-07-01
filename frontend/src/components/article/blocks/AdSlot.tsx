"use client";

import { useRouter } from "next/navigation";
import { useCallback } from "react";
import { track } from "@/lib/analytics";
import { useImpression } from "@/hooks/useImpression";
import { gamEnabled } from "@/lib/ads";
import { GamAdSlot } from "@/components/ads/GamAdSlot";
import { useT } from "@/lib/i18n/provider";

/**
 * In-content ad. Serves a real Google Ad Manager unit when configured
 * (NEXT_PUBLIC_GAM_NETWORK_CODE), otherwise shows the design placeholder. Hidden
 * for Pro readers (gated by the caller); fires one GA impression per mount.
 */
export function AdSlot({ slot = "in-content" }: { slot?: string }) {
  const router = useRouter();
  const t = useT();
  const onImpression = useCallback(
    () => track("ad_impression", { slot }),
    [slot],
  );
  const ref = useImpression<HTMLDivElement>(onImpression);

  return (
    <div
      ref={ref}
      onClick={() => track("ad_click", { slot })}
      className={`relative my-9 rounded-[14px] bg-[color:var(--faf)] p-[18px] text-center ${
        gamEnabled
          ? "border border-border"
          : "border border-dashed border-[color:var(--d8)]"
      }`}
    >
      <span className="absolute left-[14px] top-[10px] text-[10.5px] font-semibold uppercase tracking-[.1em] text-[color:var(--cb5)]">
        {t("ad.label")}
      </span>
      <div className="flex min-h-[88px] flex-col items-center justify-center gap-1.5 pt-3">
        {gamEnabled ? (
          <GamAdSlot slot={slot} />
        ) : (
          <>
            <span className="font-mono text-[12.5px] text-[color:var(--ca30)]">
              {t("ad.area")}
            </span>
            <span className="text-[12px] text-[color:var(--cbd)]">
              {t("ad.embed")}
            </span>
          </>
        )}
        <button
          onClick={(e) => {
            e.stopPropagation();
            router.push("/pro");
          }}
          className="mt-0.5 cursor-pointer border-none bg-transparent text-[12.5px] font-semibold text-accent-ink underline hover:opacity-75"
        >
          {t("ad.removePro")}
        </button>
      </div>
    </div>
  );
}
