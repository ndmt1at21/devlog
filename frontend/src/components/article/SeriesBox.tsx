"use client";

import Link from "next/link";
import type { SeriesPart } from "@/lib/types";
import { track } from "@/lib/analytics";

export function SeriesBox({
  title,
  partLabel,
  parts,
}: {
  title: string;
  partLabel: string;
  parts: SeriesPart[];
}) {
  return (
    <div
      className="mt-7 rounded-[14px] px-[18px] pb-[14px] pt-[18px]"
      style={{
        border: "1px solid color-mix(in srgb, var(--accent) 22%, transparent)",
        background: "color-mix(in srgb, var(--accent) 6%, transparent)",
      }}
    >
      <div className="mb-[3px] text-[11.5px] font-bold uppercase tracking-[.09em] text-accent-ink">
        Series · {partLabel}
      </div>
      <div className="mb-3 text-[17px] font-bold tracking-[-.01em] text-text">
        {title}
      </div>
      <div className="flex flex-col gap-0.5">
        {parts.map((pt) => {
          const inner = (
            <>
              <span className="avatar-accent h-6 w-6 flex-none text-[12.5px]">
                {pt.part}
              </span>
              <span className="flex-1">{pt.ptitle}</span>
              {pt.pLocked && (
                <span className="ml-auto text-[12.5px] text-faint">🔒</span>
              )}
            </>
          );
          const base =
            "flex items-center gap-2.5 rounded-[9px] px-2.5 py-2 text-left text-[14px] transition-colors";

          if (pt.isCurrent) {
            return (
              <div
                key={pt.id}
                className={`${base} font-bold text-text`}
                style={{
                  background:
                    "color-mix(in srgb, var(--accent) 12%, transparent)",
                }}
              >
                {inner}
              </div>
            );
          }
          return (
            <Link
              key={pt.id}
              href={`/articles/${pt.id}`}
              onClick={() =>
                track("select_article", {
                  slug: pt.id,
                  list: "series",
                  title: pt.ptitle,
                })
              }
              className={`${base} font-medium text-body no-underline hover:bg-[color:color-mix(in_srgb,var(--accent)_9%,transparent)]`}
            >
              {inner}
            </Link>
          );
        })}
      </div>
    </div>
  );
}
