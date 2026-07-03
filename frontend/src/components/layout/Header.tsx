"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { useCoffee } from "@/components/coffee/CoffeeModal";
import { donateEnabled } from "@/lib/features";
import { useT } from "@/lib/i18n/provider";
import { SITE_NAME } from "@/lib/seo";
import { AccountMenu } from "./AccountMenu";
import { SearchBox } from "./SearchBox";

export function Header() {
  const { open: openCoffee } = useCoffee();
  const t = useT();
  const [searchExpanded, setSearchExpanded] = useState(false);

  return (
    <header className="sticky top-0 z-20 border-b border-border bg-[var(--header-bg)] backdrop-blur-[10px] backdrop-saturate-[140%]">
      <div className="relative mx-auto flex max-w-[1120px] items-center justify-between gap-4 px-6 py-[13px]">
        {/* Logo (PNG wordmark — the SVG variant draws with a non-embedded font) */}
        <Link
          href="/"
          className="flex flex-none items-center no-underline"
        >
          <Image
            src="/logo.png"
            alt={SITE_NAME}
            width={64}
            height={28}
            priority
            className="logo-light"
          />
          <Image
            src="/logo-vang.png"
            alt={SITE_NAME}
            width={64}
            height={28}
            priority
            className="logo-dark"
          />
        </Link>

        {/* Desktop search */}
        <div
          role="search"
          className="relative hidden max-w-[420px] flex-1 md:block"
        >
          <SearchBox />
        </div>

        {/* Actions */}
        <div className="flex flex-none items-center gap-2.5">
          <button
            onClick={() => setSearchExpanded(true)}
            title={t("header.search")}
            aria-label={t("header.search")}
            className="flex h-[38px] w-[38px] items-center justify-center rounded-[9px] border border-border2 bg-surface text-[18px] leading-none text-strong transition-colors hover:border-hover md:hidden"
          >
            ⌕
          </button>
          {donateEnabled && (
            <button
              onClick={openCoffee}
              title={t("header.coffee")}
              aria-label={t("header.coffee")}
              className="flex h-[38px] w-[38px] items-center justify-center rounded-[9px] border border-border2 bg-surface transition-colors hover:border-hover"
            >
              <Image
                src="/buy-me-coffee-icon.png"
                alt=""
                width={20}
                height={20}
              />
            </button>
          )}
          <AccountMenu />
        </div>

        {/* Mobile expanded search overlay */}
        {searchExpanded && (
          <div
            role="search"
            className="absolute inset-0 z-[6] flex items-center gap-2 bg-[var(--header-bg)] px-6 backdrop-blur-[10px] backdrop-saturate-[140%] animate-fade-up"
          >
            <div className="relative flex-1">
              <SearchBox
                large
                autoFocus
                onNavigate={() => setSearchExpanded(false)}
              />
            </div>
            <button
              onClick={() => setSearchExpanded(false)}
              aria-label={t("common.cancel")}
              className="flex-none cursor-pointer border-none bg-transparent px-1.5 py-2 text-[15px] font-semibold text-strong hover:text-text"
            >
              {t("common.cancel")}
            </button>
          </div>
        )}
      </div>
    </header>
  );
}
