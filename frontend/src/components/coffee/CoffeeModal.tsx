"use client";

// Donate flow: PayPal (a button that opens the configured PayPal link) and MoMo
// (a QR image to scan, plus an optional deep link). Everything is configured via
// env (see src/lib/features.ts) — no backend calls. Tabs render only for the
// channels that are configured; the trigger is hidden entirely when neither is
// (see donateEnabled in Header).

import Image from "next/image";
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";
import {
  PAYPAL_DONATE_URL,
  MOMO_QR_URL,
  MOMO_DONATE_URL,
  hasPaypal,
  hasMomo,
} from "@/lib/features";

interface CoffeeContextValue {
  open: () => void;
}
const CoffeeContext = createContext<CoffeeContextValue | null>(null);

type Method = "paypal" | "momo";
// Prefer PayPal when available, otherwise fall back to MoMo.
const DEFAULT_METHOD: Method = hasPaypal ? "paypal" : "momo";
const METHODS: Array<{ key: Method; label: string; show: boolean }> = [
  { key: "paypal", label: "PayPal", show: hasPaypal },
  { key: "momo", label: "MoMo", show: hasMomo },
];

export function CoffeeProvider({ children }: { children: ReactNode }) {
  const t = useT();
  const [isOpen, setIsOpen] = useState(false);
  const [method, setMethod] = useState<Method>(DEFAULT_METHOD);

  const open = useCallback(() => {
    setMethod(DEFAULT_METHOD);
    setIsOpen(true);
    track("coffee_open", {});
  }, []);
  const close = useCallback(() => setIsOpen(false), []);

  const value = useMemo<CoffeeContextValue>(() => ({ open }), [open]);
  const tabs = METHODS.filter((m) => m.show);

  return (
    <CoffeeContext.Provider value={value}>
      {children}
      {isOpen && (
        <div
          onClick={close}
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 p-6 animate-fade-up"
        >
          <div
            role="dialog"
            aria-modal="true"
            aria-label={t("coffee.title")}
            onClick={(e) => e.stopPropagation()}
            className="w-full max-w-[400px] rounded-[18px] border border-border bg-surface p-6 shadow-[0_30px_80px_-30px_rgba(0,0,0,.6)]"
          >
            <div className="mb-1 flex items-start justify-between gap-3">
              <h3 className="m-0 text-[20px] font-extrabold tracking-[-.02em] text-text">
                {t("coffee.title")}
              </h3>
              <button
                onClick={close}
                aria-label={t("common.close")}
                className="flex-none cursor-pointer border-none bg-transparent p-0.5 text-[18px] leading-none text-faint hover:text-text"
              >
                ✕
              </button>
            </div>
            <p className="mb-[18px] mt-0 text-sm leading-[1.55] text-c5b">
              {t("coffee.subtitle")}
            </p>

            {tabs.length > 1 && (
              <div
                role="tablist"
                aria-label={t("coffee.method")}
                className="mb-4 flex gap-2"
              >
                {tabs.map((m) => {
                  const active = m.key === method;
                  return (
                    <button
                      key={m.key}
                      type="button"
                      role="tab"
                      aria-selected={active}
                      onClick={() => setMethod(m.key)}
                      className="flex-1 cursor-pointer rounded-[9px] border py-[10px] text-[14px] font-semibold transition-all"
                      style={{
                        borderColor: active
                          ? "var(--accent-ink)"
                          : "var(--border-2)",
                        background: active
                          ? "color-mix(in srgb, var(--accent) 10%, transparent)"
                          : "var(--surface)",
                        color: active ? "var(--accent-ink)" : "var(--c5b)",
                      }}
                    >
                      {m.label}
                    </button>
                  );
                })}
              </div>
            )}

            {method === "paypal" && hasPaypal && (
              <div className="py-1 text-center">
                <p className="mb-4 mt-0 text-[13.5px] leading-[1.5] text-muted">
                  {t("coffee.paypalHint")}
                </p>
                <a
                  href={PAYPAL_DONATE_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  onClick={() => track("coffee_donate", { method: "paypal" })}
                  className="btn-accent flex w-full items-center justify-center gap-[7px] py-[13px] text-[15px] no-underline"
                >
                  {t("coffee.paypalButton")}
                  <span aria-hidden="true">↗</span>
                </a>
              </div>
            )}

            {method === "momo" && hasMomo && (
              <div className="py-1 text-center">
                <Image
                  src={MOMO_QR_URL}
                  alt={t("coffee.momoQrAlt")}
                  width={200}
                  height={200}
                  unoptimized
                  className="mx-auto mb-3 h-auto w-[200px] rounded-[14px] border border-border2 bg-white p-2"
                />
                <p className="m-0 mb-4 text-[13.5px] leading-[1.5] text-muted">
                  {t("coffee.momoScan")}
                </p>
                {MOMO_DONATE_URL && (
                  <a
                    href={MOMO_DONATE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={() => track("coffee_donate", { method: "momo" })}
                    className="btn-outline inline-flex items-center justify-center gap-[7px] px-[22px] py-[10px] text-[14.5px] no-underline"
                  >
                    {t("coffee.momoButton")}
                    <span aria-hidden="true">↗</span>
                  </a>
                )}
              </div>
            )}

            <p className="mt-4 text-center text-[12px] text-faint">
              {t("coffee.donateNote")}
            </p>
          </div>
        </div>
      )}
    </CoffeeContext.Provider>
  );
}

export function useCoffee(): CoffeeContextValue {
  const ctx = useContext(CoffeeContext);
  if (!ctx) throw new Error("useCoffee must be used within <CoffeeProvider>");
  return ctx;
}
