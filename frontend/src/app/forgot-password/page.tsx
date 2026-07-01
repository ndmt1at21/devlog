import type { Metadata } from "next";
import { ForgotForm } from "@/components/auth/ForgotForm";

export const metadata: Metadata = {
  title: "Quên mật khẩu",
  robots: { index: false, follow: true },
};

export default function ForgotPasswordPage() {
  return <ForgotForm />;
}
