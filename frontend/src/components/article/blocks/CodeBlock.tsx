"use client";

import { useState } from "react";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

export function CodeBlock({
  lang,
  code,
  html,
  slug,
}: {
  lang?: string;
  code: string;
  html?: string;
  slug: string;
}) {
  const [copied, setCopied] = useState(false);
  const t = useT();

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      track("copy_code", { language: lang ?? "", slug });
      setTimeout(() => setCopied(false), 1600);
    } catch {
      /* clipboard unavailable */
    }
  };

  return (
    <div className="my-[26px] overflow-hidden rounded-xl border border-[#2a2e35] bg-[#1c1f24]">
      <div className="flex items-center justify-between border-b border-[#2a2e35] bg-[#23272e] px-[14px] py-[9px]">
        <span className="font-mono text-[11.5px] uppercase tracking-[.06em] text-[#8b9099]">
          {lang}
        </span>
        <button
          onClick={copy}
          title={lang ? `${t("code.copy")} (${lang})` : t("code.copy")}
          aria-live="polite"
          className="cursor-pointer rounded-[7px] border border-[#3a3f47] bg-transparent px-[11px] py-1 text-[12px] font-medium text-[#aeb3bb] transition-all hover:border-[#4d535c] hover:text-[#e6e8eb]"
        >
          {copied ? t("code.copied") : t("code.copy")}
        </button>
      </div>
      {html ? (
        <div
          className="code-shiki"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      ) : (
        <pre className="m-0 overflow-x-auto whitespace-pre px-[18px] py-4 font-mono text-[13.5px] leading-[1.7] text-[#dfe3e8] [tab-size:4]">
          {code}
        </pre>
      )}
    </div>
  );
}
