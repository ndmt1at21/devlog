import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { ProContent } from "@/components/pro/ProContent";
import { proEnabled } from "@/lib/features";

export const metadata: Metadata = {
  title: "devnote Pro",
  description: "Đọc không giới hạn, không quảng cáo. Ủng hộ devnote với gói Pro.",
};

export default function ProPage() {
  // PRO subscription can be turned off via NEXT_PUBLIC_PRO_ENABLED.
  if (!proEnabled) notFound();
  return <ProContent />;
}
