"use client";

import Link from "next/link";
import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { useT } from "@/lib/i18n/provider";

export function ForgotForm() {
  const t = useT();
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [sent, setSent] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (!/^\S+@\S+\.\S+$/.test(email.trim())) {
      setError(t("auth.forgot.errEmail"));
      return;
    }
    setBusy(true);
    try {
      await api.forgotPassword({ email: email.trim() });
      setSent(true);
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : t("auth.forgot.errFailed"),
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-[420px] px-6 pb-24 pt-16">
      <h1 className="m-0 mb-2 text-center text-[30px] font-extrabold tracking-[-.03em] text-text">
        {t("auth.forgot.title")}
      </h1>
      <p className="m-0 mb-[30px] text-center text-[15px] text-muted">
        {t("auth.forgot.sub")}
      </p>
      <div className="rounded-2xl border border-border bg-surface p-7">
        {sent ? (
          <div className="py-2 text-center">
            <div aria-hidden="true" className="mb-3 text-[40px]">📬</div>
            <h3 className="m-0 mb-2 text-[18px] font-bold text-text">
              {t("auth.sentTitle")}
            </h3>
            <p className="m-0 text-[14.5px] leading-[1.6] text-c5b">
              {t("auth.forgot.sentBody", { email: email.trim() })}
            </p>
          </div>
        ) : (
          <form onSubmit={submit}>
            <label
              htmlFor="forgot-email"
              className="mb-[7px] block text-[13.5px] font-semibold text-strong"
            >
              {t("auth.email")}
            </label>
            <input
              id="forgot-email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t("auth.emailPlaceholder")}
              className="field mb-1.5 px-[14px] py-3 text-[15px]"
            />
            <div
              role="alert"
              className="mb-2 min-h-[20px] text-[13px] font-medium text-danger"
            >
              {error}
            </div>
            <button
              type="submit"
              disabled={busy}
              className="btn-accent w-full py-[13px] text-[15px] disabled:opacity-60"
            >
              {busy ? t("auth.forgot.submitting") : t("auth.forgot.submit")}
            </button>
          </form>
        )}
      </div>
      <p className="mt-[22px] text-center text-[14.5px] text-muted">
        <Link
          href="/login"
          className="font-semibold text-accent-ink no-underline hover:opacity-75"
        >
          {t("auth.forgot.back")}
        </Link>
      </p>
    </div>
  );
}
