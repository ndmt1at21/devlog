"use client";

import Image from "next/image";
import { useT } from "@/lib/i18n/provider";
import { SITE_NAME } from "@/lib/seo";

export function Footer() {
  const t = useT();
  return (
    <footer className="border-t border-border bg-surface">
      <div className="mx-auto flex max-w-[1120px] flex-wrap items-center justify-between gap-5 px-6 py-[34px]">
        <div className="flex items-center gap-2.5">
          <Image
            src="/logo.png"
            alt={SITE_NAME}
            width={54}
            height={24}
            className="logo-light"
          />
          <Image
            src="/logo-vang.png"
            alt={SITE_NAME}
            width={54}
            height={24}
            className="logo-dark"
          />
        </div>
        <span className="text-[13px] text-faint">{t("footer.copyright")}</span>
      </div>
    </footer>
  );
}
