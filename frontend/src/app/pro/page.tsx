import type { Metadata } from "next";
import { ProContent } from "@/components/pro/ProContent";

export const metadata: Metadata = {
  title: "devnote Pro",
  description: "Đọc không giới hạn, không quảng cáo. Ủng hộ devnote với gói Pro.",
};

export default function ProPage() {
  return <ProContent />;
}
