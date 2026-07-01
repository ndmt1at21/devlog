"use client";

import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { useSearch } from "@/lib/search";
import { useCoffee } from "@/components/coffee/CoffeeModal";
import { useT } from "@/lib/i18n/provider";
import { AccountMenu } from "./AccountMenu";

export function Header() {
  const { query, setQuery } = useSearch();
  const { open: openCoffee } = useCoffee();
  const t = useT();
  const [searchExpanded, setSearchExpanded] = useState(false);

  return (
    <header className="sticky top-0 z-20 border-b border-border bg-[var(--header-bg)] backdrop-blur-[10px] backdrop-saturate-[140%]">
      <div className="relative mx-auto flex max-w-[1120px] items-center justify-between gap-4 px-6 py-[13px]">
        {/* Logo */}
        <Link
          href="/"
          className="flex flex-none items-center gap-[9px] no-underline"
        >
          <span className="flex h-8 w-8 items-center justify-center rounded-[9px] bg-accent text-[15px] font-bold text-on-accent">
            {"{ }"}
          </span>
          <span className="text-[17px] font-extrabold tracking-[-.02em] text-text">
            devnote
          </span>
        </Link>

        {/* Desktop search */}
        <div
          role="search"
          className="relative hidden max-w-[420px] flex-1 md:block"
        >
          <span
            aria-hidden="true"
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[15px] text-faint"
          >
            ⌕
          </span>
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("header.searchPlaceholder")}
            aria-label={t("header.search")}
            className="field py-[9px] pl-[33px] pr-[14px] text-sm"
          />
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
          <AccountMenu />
        </div>

        {/* Mobile expanded search overlay */}
        {searchExpanded && (
          <div
            role="search"
            className="absolute inset-0 z-[6] flex items-center gap-2 bg-[var(--header-bg)] px-6 backdrop-blur-[10px] backdrop-saturate-[140%] animate-fade-up"
          >
            <span aria-hidden="true" className="flex-none text-[19px] text-faint">
              ⌕
            </span>
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t("header.searchPlaceholder")}
              aria-label={t("header.search")}
              className="field flex-1 px-3 py-[10px] text-[15px]"
            />
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
