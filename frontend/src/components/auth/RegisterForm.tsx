"use client";

import Link from "next/link";
import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { useT } from "@/lib/i18n/provider";
import { track } from "@/lib/analytics";
import { GoogleButton, OrDivider } from "./GoogleButton";

export function RegisterForm() {
  const t = useT();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [sent, setSent] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (!name.trim() || !email.trim() || !password) {
      setError(t("auth.register.errAllFields"));
      return;
    }
    if (password.length < 6) {
      setError(t("auth.register.errWeakPass"));
      return;
    }
    setBusy(true);
    try {
      const res = await api.register({
        name: name.trim(),
        email: email.trim(),
        password,
      });
      track("sign_up", { method: "password" });
      setSent(res.message);
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : t("auth.register.errFailed"),
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-[420px] px-6 pb-24 pt-16">
      <h1 className="m-0 mb-2 text-center text-[30px] font-extrabold tracking-[-.03em] text-text">
        {t("auth.register.title")}
      </h1>
      <p className="m-0 mb-[30px] text-center text-[15px] text-muted">
        {t("auth.register.sub")}
      </p>

      {sent ? (
        <div className="rounded-2xl border border-border bg-surface p-7 text-center">
          <div aria-hidden="true" className="mb-3 text-[40px]">📬</div>
          <h3 className="m-0 mb-2 text-[18px] font-bold text-text">
            {t("auth.sentTitle")}
          </h3>
          <p className="m-0 text-[14.5px] leading-[1.6] text-c5b">{sent}</p>
        </div>
      ) : (
        <form
          onSubmit={submit}
          className="rounded-2xl border border-border bg-surface p-7"
        >
          <GoogleButton label={t("auth.googleRegister")} event="sign_up" />
          <OrDivider />
          <label
            htmlFor="register-name"
            className="mb-[7px] block text-[13.5px] font-semibold text-strong"
          >
            {t("auth.displayName")}
          </label>
          <input
            id="register-name"
            autoComplete="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("auth.namePlaceholder")}
            className="field mb-[18px] px-[14px] py-3 text-[15px]"
          />
          <label
            htmlFor="register-email"
            className="mb-[7px] block text-[13.5px] font-semibold text-strong"
          >
            {t("auth.email")}
          </label>
          <input
            id="register-email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder={t("auth.emailPlaceholder")}
            className="field mb-[18px] px-[14px] py-3 text-[15px]"
          />
          <label
            htmlFor="register-password"
            className="mb-[7px] block text-[13.5px] font-semibold text-strong"
          >
            {t("auth.password")}
          </label>
          <input
            id="register-password"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder={t("auth.regPassPlaceholder")}
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
            {busy ? t("auth.register.submitting") : t("auth.register.submit")}
          </button>
        </form>
      )}

      <p className="mt-[22px] text-center text-[14.5px] text-muted">
        {t("auth.register.have")}{" "}
        <Link
          href="/login"
          className="font-semibold text-accent-ink no-underline hover:opacity-75"
        >
          {t("auth.register.login")}
        </Link>
      </p>
    </div>
  );
}
