"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useT } from "@/lib/i18n/provider";
import { track } from "@/lib/analytics";
import type { Plan } from "@/lib/types";

const PLAN_VALUE: Record<string, number> = { month: 39000, year: 299000 };

export function ProContent() {
  const router = useRouter();
  const { user, premium, refresh } = useAuth();
  const t = useT();
  const [plans, setPlans] = useState<Plan[]>([]);
  const [selected, setSelected] = useState<"month" | "year">("year");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const BENEFITS = [
    t("pro.benefit1"),
    t("pro.benefit2"),
    t("pro.benefit3"),
  ];

  useEffect(() => {
    api.plans().then(setPlans).catch(() => setPlans([]));
  }, []);

  const subscribe = async () => {
    setError("");
    if (!user) {
      router.push("/login");
      return;
    }
    setBusy(true);
    try {
      await api.subscribe(selected);
      track("subscribe_pro", {
        plan: selected,
        value: PLAN_VALUE[selected],
        currency: "VND",
      });
      await refresh();
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Không kích hoạt được Pro.",
      );
    } finally {
      setBusy(false);
    }
  };

  if (premium) {
    return (
      <div className="mx-auto max-w-[520px] px-6 pb-24 pt-[60px]">
        <div className="py-[30px] text-center">
          <div className="mb-3.5 text-[48px]">🎉</div>
          <h1 className="m-0 mb-2.5 text-[28px] font-extrabold tracking-[-.03em] text-text">
            {t("pro.successTitle")}
          </h1>
          <p className="mx-auto mb-[26px] max-w-[360px] text-[15px] leading-[1.6] text-c5b">
            {t("pro.successBody")}
          </p>
          <button
            onClick={() => router.push("/")}
            className="btn-accent px-[26px] py-3 text-[15px]"
          >
            {t("pro.startReading")}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-[520px] px-6 pb-24 pt-[60px]">
      <div className="mb-[30px] text-center">
        <div
          className="mb-4 inline-flex items-center gap-[7px] rounded-full px-3.5 py-1.5 text-[13px] font-bold text-accent-ink"
          style={{ background: "color-mix(in srgb, var(--accent) 12%, transparent)" }}
        >
          {t("pro.badge")}
        </div>
        <h1 className="m-0 mb-2.5 text-[31px] font-extrabold tracking-[-.03em] text-text">
          {t("pro.heading")}
        </h1>
        <p className="mx-auto max-w-[400px] text-[15.5px] leading-[1.6] text-c5b">
          {t("pro.sub")}
        </p>
      </div>

      <div className="rounded-[18px] border border-border bg-surface p-[26px]">
        <div className="mb-6 flex flex-col gap-3">
          {BENEFITS.map((b) => (
            <div key={b} className="flex items-center gap-3 text-[15px] text-strong">
              <span className="avatar-accent h-[22px] w-[22px] text-[12px]">✓</span>
              {b}
            </div>
          ))}
        </div>

        <div className="mb-5 flex gap-[11px]">
          {plans.map((p) => {
            const active = p.key === selected;
            return (
              <button
                key={p.key}
                onClick={() => {
                  setSelected(p.key);
                  track("select_pro_plan", {
                    plan: p.key,
                    price: p.price,
                    value: PLAN_VALUE[p.key],
                    currency: "VND",
                  });
                }}
                className="flex flex-1 flex-col gap-1 rounded-[14px] border px-4 py-3.5 text-left transition-all"
                style={{
                  borderColor: active ? "var(--accent-ink)" : "var(--border-2)",
                  background: active
                    ? "color-mix(in srgb, var(--accent) 8%, transparent)"
                    : "var(--surface)",
                }}
              >
                <span className="text-[13px] font-semibold text-c5b">
                  {p.name}
                </span>
                <span className="text-[23px] font-extrabold tracking-[-.02em] text-text">
                  {p.price}
                </span>
                <span className="text-[12px] text-muted">{p.note}</span>
              </button>
            );
          })}
        </div>

        {error && (
          <div className="mb-3 text-center text-[13px] font-medium text-danger">
            {error}
          </div>
        )}

        <button
          onClick={subscribe}
          disabled={busy}
          className="btn-accent w-full rounded-[11px] py-3.5 text-[15.5px] disabled:opacity-60"
        >
          {busy ? t("pro.subscribing") : t("pro.subscribe")}
        </button>
        <p className="mt-[13px] text-center text-[12.5px] text-faint">
          {t("pro.demo")}
        </p>
      </div>

      <p className="mt-5 text-center text-[14px] text-muted">
        <button
          onClick={() => router.push("/")}
          className="cursor-pointer border-none bg-transparent p-0 text-[14px] font-semibold text-accent-ink hover:opacity-75"
        >
          {t("common.back")}
        </button>
      </p>
    </div>
  );
}
