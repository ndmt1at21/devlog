"use client";

import { useEffect } from "react";
import { useT } from "@/lib/i18n/provider";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const t = useT();
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="mx-auto max-w-[520px] px-6 pb-24 pt-24 text-center">
      <div className="mb-4 text-[48px]">⚠️</div>
      <h1 className="m-0 mb-3 text-[24px] font-extrabold tracking-[-.02em] text-text">
        {t("error.title")}
      </h1>
      <p className="mx-auto mb-7 max-w-[360px] text-[15px] leading-[1.6] text-c5b">
        {t("error.body")}
      </p>
      <button onClick={reset} className="btn-accent px-[26px] py-3 text-[15px]">
        {t("error.retry")}
      </button>
    </div>
  );
}
