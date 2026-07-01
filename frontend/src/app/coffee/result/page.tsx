"use client";

// Landing page for real Stripe/MoMo redirects (demo flow never reaches here).
// It polls the order status until the provider confirms payment.

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useT } from "@/lib/i18n/provider";

function Result() {
  const t = useT();
  const params = useSearchParams();
  const order = params.get("order");
  const canceled = params.get("canceled");
  const [status, setStatus] = useState<"pending" | "completed" | "failed">(
    canceled ? "failed" : "pending",
  );

  useEffect(() => {
    if (!order || canceled) return;
    let tries = 0;
    const tick = async () => {
      try {
        const res = await api.coffeeStatus(order);
        if (res.status === "completed") return setStatus("completed");
        if (res.status === "failed") return setStatus("failed");
      } catch {
        /* keep polling */
      }
    };
    void tick();
    const id = setInterval(() => {
      tries += 1;
      if (tries > 20) {
        clearInterval(id);
        return;
      }
      void tick();
    }, 2000);
    return () => clearInterval(id);
  }, [order, canceled]);

  const view = {
    pending: {
      icon: "⏳",
      title: t("coffeeResult.pendingTitle"),
      sub: t("coffeeResult.pendingSub"),
    },
    completed: {
      icon: "🙏",
      title: t("coffeeResult.completedTitle"),
      sub: t("coffeeResult.completedSub"),
    },
    failed: {
      icon: "😕",
      title: t("coffeeResult.failedTitle"),
      sub: t("coffeeResult.failedSub"),
    },
  }[status];

  return (
    <div className="mx-auto max-w-[520px] px-6 pb-24 pt-24 text-center">
      <div className="mb-4 text-[48px]">{view.icon}</div>
      <h1 className="m-0 mb-3 text-[24px] font-extrabold tracking-[-.02em] text-text">
        {view.title}
      </h1>
      <p className="mx-auto mb-7 max-w-[360px] text-[15px] leading-[1.6] text-c5b">
        {view.sub}
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

export default function CoffeeResultPage() {
  return (
    <Suspense fallback={null}>
      <Result />
    </Suspense>
  );
}
