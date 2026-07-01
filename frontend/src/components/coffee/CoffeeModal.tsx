"use client";

// "Buy me a coffee" flow: amount → pay (card/MoMo) → done. When no real payment
// provider is configured the backend returns {demo:true} and we jump straight to
// the thank-you step (matches the mockup's "bản demo — không phát sinh thanh toán").

import Image from "next/image";
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { api } from "@/lib/api";
import { track } from "@/lib/analytics";
import { formatVND } from "@/lib/format";
import { useT } from "@/lib/i18n/provider";

interface CoffeeContextValue {
  open: () => void;
}
const CoffeeContext = createContext<CoffeeContextValue | null>(null);

const OPTIONS = [
  { amt: 25000, qty: 1 },
  { amt: 75000, qty: 3 },
  { amt: 125000, qty: 5 },
] as const;

type Step = "amount" | "pay" | "done";
type Method = "card" | "momo";

export function CoffeeProvider({ children }: { children: ReactNode }) {
  const t = useT();
  const [isOpen, setIsOpen] = useState(false);
  const [step, setStep] = useState<Step>("amount");
  const [amount, setAmount] = useState<number>(75000);
  const [name, setName] = useState("");
  const [method, setMethod] = useState<Method>("card");
  const [pay, setPay] = useState({ card: "", exp: "", cvc: "" });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const open = useCallback(() => {
    setStep("amount");
    setError("");
    setIsOpen(true);
    track("coffee_open", {});
  }, []);

  const close = useCallback(() => setIsOpen(false), []);
  const qty = amount / 25000;

  const submitPay = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");
      setBusy(true);
      try {
        const res = await api.coffeeCheckout({ amount, method, name });
        track("coffee_donate", {
          amount,
          method,
          value: amount,
          currency: "VND",
        });
        if (res.redirectUrl) {
          window.location.href = res.redirectUrl; // Stripe Checkout
          return;
        }
        if (res.payUrl) {
          window.location.href = res.payUrl; // MoMo
          return;
        }
        setStep("done"); // demo flow
      } catch {
        setError(t("coffee.errFailed"));
      } finally {
        setBusy(false);
      }
    },
    [amount, method, name, t],
  );

  const value = useMemo<CoffeeContextValue>(() => ({ open }), [open]);

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
            {step === "amount" && (
              <>
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
                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    setStep("pay");
                  }}
                >
                  <div className="mb-[14px] flex gap-[10px]">
                    {OPTIONS.map((o) => {
                      const active = o.amt === amount;
                      return (
                        <button
                          key={o.amt}
                          type="button"
                          onClick={() => setAmount(o.amt)}
                          aria-pressed={active}
                          aria-label={formatVND(o.amt)}
                          className="flex flex-1 flex-col items-center gap-1.5 rounded-[11px] border py-[13px] transition-all"
                          style={{
                            borderColor: active
                              ? "var(--accent-ink)"
                              : "var(--border-2)",
                            background: active
                              ? "color-mix(in srgb, var(--accent) 10%, transparent)"
                              : "var(--surface)",
                          }}
                        >
                          <Image src="/buy-me-coffee-icon.png" alt="" width={20} height={20} />
                          <span className="text-[13px] font-bold text-text">
                            {formatVND(o.amt)}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                  <input
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder={t("coffee.namePlaceholder")}
                    aria-label={t("coffee.namePlaceholder")}
                    className="field mb-[14px] px-[14px] py-[11px] text-[14.5px]"
                  />
                  <button
                    type="submit"
                    className="btn-accent w-full py-3 text-[15px]"
                  >
                    {t("coffee.continue", { amount: formatVND(amount) })}
                  </button>
                </form>
              </>
            )}

            {step === "pay" && (
              <>
                <div className="mb-4 flex items-start justify-between gap-3">
                  <div className="flex items-center gap-[10px]">
                    <button
                      onClick={() => setStep("amount")}
                      aria-label={t("common.back")}
                      className="flex-none cursor-pointer border-none bg-transparent p-0.5 text-[18px] leading-none text-faint hover:text-text"
                    >
                      ←
                    </button>
                    <h3 className="m-0 text-[20px] font-extrabold tracking-[-.02em] text-text">
                      {t("coffee.pay")}
                    </h3>
                  </div>
                  <button
                    onClick={close}
                    aria-label={t("common.close")}
                    className="flex-none cursor-pointer border-none bg-transparent p-0.5 text-[18px] leading-none text-faint hover:text-text"
                  >
                    ✕
                  </button>
                </div>

                <div
                  className="mb-[18px] flex items-center justify-between rounded-[11px] px-[15px] py-[13px]"
                  style={{
                    background:
                      "color-mix(in srgb, var(--accent) 8%, var(--surface))",
                    border:
                      "1px solid color-mix(in srgb, var(--accent) 22%, transparent)",
                  }}
                >
                  <span className="flex items-center gap-[9px] text-sm font-semibold text-strong">
                    <Image src="/buy-me-coffee-icon.png" alt="" width={20} height={20} />
                    {t("coffee.qty", { count: qty })}
                  </span>
                  <span className="text-[17px] font-extrabold text-text">
                    {formatVND(amount)}
                  </span>
                </div>

                <div className="mb-4 flex gap-2">
                  {(
                    [
                      { key: "card", label: t("coffee.methodCard") },
                      { key: "momo", label: "MoMo" },
                    ] as const
                  ).map((m) => {
                    const active = m.key === method;
                    return (
                      <button
                        key={m.key}
                        type="button"
                        onClick={() => setMethod(m.key)}
                        aria-pressed={active}
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

                <form onSubmit={submitPay}>
                  {method === "card" ? (
                    <>
                      <label
                        htmlFor="coffee-card"
                        className="mb-1.5 block text-[13px] font-semibold text-strong"
                      >
                        {t("coffee.cardNumber")}
                      </label>
                      <input
                        id="coffee-card"
                        value={pay.card}
                        onChange={(e) =>
                          setPay((p) => ({ ...p, card: e.target.value }))
                        }
                        inputMode="numeric"
                        placeholder="4242 4242 4242 4242"
                        className="field mb-3 px-[14px] py-[11px] text-[14.5px] tracking-[.04em]"
                      />
                      <div className="mb-[14px] flex gap-[10px]">
                        <div className="flex-1">
                          <label
                            htmlFor="coffee-exp"
                            className="mb-1.5 block text-[13px] font-semibold text-strong"
                          >
                            {t("coffee.expiry")}
                          </label>
                          <input
                            id="coffee-exp"
                            value={pay.exp}
                            onChange={(e) =>
                              setPay((p) => ({ ...p, exp: e.target.value }))
                            }
                            placeholder="MM/YY"
                            className="field px-[14px] py-[11px] text-[14.5px]"
                          />
                        </div>
                        <div className="w-[110px]">
                          <label
                            htmlFor="coffee-cvc"
                            className="mb-1.5 block text-[13px] font-semibold text-strong"
                          >
                            CVC
                          </label>
                          <input
                            id="coffee-cvc"
                            value={pay.cvc}
                            onChange={(e) =>
                              setPay((p) => ({ ...p, cvc: e.target.value }))
                            }
                            inputMode="numeric"
                            placeholder="123"
                            className="field px-[14px] py-[11px] text-[14.5px]"
                          />
                        </div>
                      </div>
                    </>
                  ) : (
                    <div className="py-[14px] pb-[18px] text-center">
                      <div
                        className="mx-auto mb-3 h-[140px] w-[140px] rounded-[14px] border-[6px] border-surface"
                        style={{
                          background:
                            "repeating-conic-gradient(var(--text) 0% 25%, var(--surface) 0% 50%) 50% / 16px 16px",
                          boxShadow: "0 0 0 1px var(--border-2)",
                        }}
                      />
                      <p className="m-0 text-[13.5px] leading-[1.5] text-muted">
                        {t("coffee.momoScan")}
                      </p>
                    </div>
                  )}
                  <div className="mb-2 min-h-[18px] text-[13px] font-medium text-danger">
                    {error}
                  </div>
                  <button
                    type="submit"
                    disabled={busy}
                    className="btn-accent flex w-full items-center justify-center gap-[7px] py-[13px] text-[15px] disabled:opacity-60"
                  >
                    <span aria-hidden="true" className="text-[13px]">🔒</span>
                    {t("coffee.payAmount", { amount: formatVND(amount) })}
                  </button>
                  <p className="mt-3 text-center text-[12px] text-faint">
                    {t("coffee.demo")}
                  </p>
                </form>
              </>
            )}

            {step === "done" && (
              <div className="py-[14px] pb-1.5 text-center">
                <div aria-hidden="true" className="mb-3 text-[44px]">🙏</div>
                <h3 className="m-0 mb-2 text-[20px] font-extrabold tracking-[-.02em] text-text">
                  {t("coffee.doneTitle")}
                </h3>
                <p className="mb-[22px] mt-0 text-[14.5px] leading-[1.6] text-c5b">
                  {t("coffee.doneBody", { amount: formatVND(amount) })}
                </p>
                <button
                  onClick={close}
                  className="btn-outline px-[22px] py-[10px] text-[14.5px]"
                >
                  {t("common.close")}
                </button>
              </div>
            )}
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
