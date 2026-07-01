import Link from "next/link";
import { getT } from "@/lib/i18n/server";

export default async function NotFound() {
  const t = await getT();
  return (
    <div className="mx-auto max-w-[520px] px-6 pb-24 pt-24 text-center">
      <div className="mb-4 font-mono text-[64px] font-extrabold text-accent">
        {"{ }"}
      </div>
      <h1 className="m-0 mb-3 text-[26px] font-extrabold tracking-[-.02em] text-text">
        {t("notFound.title")}
      </h1>
      <p className="mx-auto mb-7 max-w-[360px] text-[15px] leading-[1.6] text-c5b">
        {t("notFound.body")}
      </p>
      <Link
        href="/"
        className="btn-accent inline-block px-[26px] py-3 text-[15px] no-underline"
      >
        {t("common.home")}
      </Link>
    </div>
  );
}
