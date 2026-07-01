"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useT } from "@/lib/i18n/provider";
import { track } from "@/lib/analytics";
import { GoogleButton, OrDivider } from "./GoogleButton";

// Maps the ?error= query codes (set by the BFF google flow) to dict keys.
const ERROR_KEYS: Record<string, string> = {
  auth_unavailable: "auth.err.authUnavailable",
  google_failed: "auth.err.googleFailed",
};

export function LoginForm({ initialError }: { initialError?: string }) {
  const router = useRouter();
  const { refresh } = useAuth();
  const t = useT();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState(
    initialError && ERROR_KEYS[initialError] ? t(ERROR_KEYS[initialError]) : "",
  );
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (!email.trim() || !password) {
      setError(t("auth.login.errEmailPass"));
      return;
    }
    setBusy(true);
    try {
      await api.login({ email: email.trim(), password });
      track("login", { method: "password" });
      await refresh();
      router.push("/");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("auth.login.errFailed"));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-[420px] px-6 pb-24 pt-16">
      <h1 className="m-0 mb-2 text-center text-[30px] font-extrabold tracking-[-.03em] text-text">
        {t("auth.login.title")}
      </h1>
      <p className="m-0 mb-[30px] text-center text-[15px] text-muted">
        {t("auth.login.sub")}
      </p>
      <form onSubmit={submit} className="rounded-2xl border border-border bg-surface p-7">
        <GoogleButton label={t("auth.googleLogin")} event="login" />
        <OrDivider />
        <label
          htmlFor="login-email"
          className="mb-[7px] block text-[13.5px] font-semibold text-strong"
        >
          {t("auth.email")}
        </label>
        <input
          id="login-email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder={t("auth.emailPlaceholder")}
          className="field mb-[18px] px-[14px] py-3 text-[15px]"
        />
        <label
          htmlFor="login-password"
          className="mb-[7px] block text-[13.5px] font-semibold text-strong"
        >
          {t("auth.password")}
        </label>
        <input
          id="login-password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="••••••••"
          className="field mb-2.5 px-[14px] py-3 text-[15px]"
        />
        <div className="mb-2 text-right">
          <Link
            href="/forgot-password"
            className="text-[13px] font-semibold text-accent-ink no-underline hover:opacity-75"
          >
            {t("auth.login.forgot")}
          </Link>
        </div>
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
          {busy ? t("auth.login.submitting") : t("auth.login.submit")}
        </button>
        <div className="mt-3.5 rounded-[10px] bg-soft px-[14px] py-[11px] text-[12.5px] leading-[1.5] text-subtle">
          {t("auth.login.demoPrefix")}{" "}
          <b className="text-strong">demo@blog.vn</b> · {t("auth.login.demoPass")}{" "}
          <b className="text-strong">123456</b>
        </div>
      </form>
      <p className="mt-[22px] text-center text-[14.5px] text-muted">
        {t("auth.login.noAccount")}{" "}
        <Link
          href="/register"
          className="font-semibold text-accent-ink no-underline hover:opacity-75"
        >
          {t("auth.login.signupNow")}
        </Link>
      </p>
    </div>
  );
}
