"use client";

import { useT } from "@/lib/i18n/provider";

export function Footer() {
  const t = useT();
  return (
    <footer className="border-t border-border bg-surface">
      <div className="mx-auto flex max-w-[1120px] flex-wrap items-center justify-between gap-5 px-6 py-[34px]">
        <div className="flex items-center gap-2.5">
          <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-accent text-[13px] font-bold text-on-accent">
            {"{ }"}
          </span>
          <span className="text-[16px] font-bold text-text">devnote</span>
          <span className="ml-1.5 text-[13.5px] text-faint">
            {t("footer.tagline")}
          </span>
        </div>
        <span className="text-[13px] text-faint">{t("footer.copyright")}</span>
      </div>
    </footer>
  );
}
