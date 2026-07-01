import type { Metadata } from "next";
import { RegisterForm } from "@/components/auth/RegisterForm";

export const metadata: Metadata = {
  title: "Đăng ký",
  robots: { index: false, follow: true },
};

export default function RegisterPage() {
  return <RegisterForm />;
}
