"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { useAuth } from "@/lib/auth";
import { useTheme } from "@/components/theme/ThemeProvider";
import { useI18n } from "@/lib/i18n/provider";
import { LOCALES, LOCALE_LABELS } from "@/lib/i18n/dictionaries";
import { track } from "@/lib/analytics";
import { proEnabled } from "@/lib/features";
import { initial } from "@/lib/format";

export function AccountMenu() {
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const { user, premium, logout } = useAuth();
  const { theme, toggle } = useTheme();
  const { locale, setLocale, t } = useI18n();
  const loggedIn = !!user;

  const go = (path: string) => {
    setOpen(false);
    router.push(path);
  };

  const item =
    "flex w-full items-center gap-[11px] rounded-[9px] px-3 py-[9px] text-left text-sm transition-colors hover:bg-hoverbg";

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={t("header.account")}
        className="flex cursor-pointer items-center gap-1.5 rounded-full border border-border2 bg-surface py-[3px] pl-[3px] pr-[9px] text-strong transition-colors hover:border-hover"
      >
        {loggedIn ? (
          <span className="avatar-accent h-[30px] w-[30px] text-[13px]">
            {initial(user.name)}
          </span>
        ) : (
          <span
            aria-hidden="true"
            className="flex h-[30px] w-[30px] items-center justify-center rounded-full bg-hoverbg text-[15px] text-c5b"
          >
            ☰
          </span>
        )}
        <span aria-hidden="true" className="text-[10px] text-faint">
          ▾
        </span>
      </button>

      {open && (
        <>
          <div
            className="fixed inset-0 z-[1]"
            aria-hidden="true"
            onClick={() => setOpen(false)}
          />
          <div
            role="menu"
            aria-label={t("header.account")}
            className="absolute right-0 top-[calc(100%+10px)] z-[2] w-[230px] rounded-[14px] border border-border bg-surface p-1.5 shadow-[0_18px_50px_-18px_rgba(0,0,0,.35)] animate-fade-up"
          >
            {loggedIn ? (
              <div className="mb-1.5 border-b border-border px-3 pb-[11px] pt-[9px]">
                <div className="text-[14px] font-bold text-text">
                  {user.name}
                </div>
                <div className="mt-0.5 truncate text-[12.5px] text-faint">
                  {user.email}
                </div>
              </div>
            ) : (
              <>
                <button
                  role="menuitem"
                  onClick={() => go("/login")}
                  className={`${item} font-medium text-strong`}
                >
                  {t("account.login")}
                </button>
                <button
                  role="menuitem"
                  onClick={() => go("/register")}
                  className={`${item} font-bold text-accent-ink`}
                >
                  {t("account.register")}
                </button>
                <div className="mx-2 my-1.5 h-px bg-border" />
              </>
            )}

            {loggedIn && user.canWrite && (
              <button
                role="menuitem"
                onClick={() => go("/articles/new")}
                className={`${item} font-medium text-strong`}
              >
                <span
                  aria-hidden="true"
                  className="w-[18px] text-center text-[15px]"
                >
                  ✍️
                </span>
                {t("account.write")}
              </button>
            )}

            {proEnabled &&
              (premium ? (
                <button
                  role="menuitem"
                  onClick={() => go("/pro")}
                  className={`${item} font-medium text-strong`}
                >
                  <span
                    aria-hidden="true"
                    className="w-[18px] text-center text-[15px] text-accent-ink"
                  >
                    ✦
                  </span>
                  {t("account.memberPro")}
                </button>
              ) : (
                <button
                  role="menuitem"
                  onClick={() => go("/pro")}
                  className={`${item} font-bold text-accent-ink`}
                >
                  <span aria-hidden="true" className="w-[18px] text-center text-[15px]">
                    ✦
                  </span>
                  {t("account.upgradePro")}
                </button>
              ))}

            <button
              role="menuitem"
              onClick={toggle}
              className={`${item} font-medium text-strong`}
            >
              <span aria-hidden="true" className="w-[18px] text-center text-[16px]">
                {theme === "dark" ? "☀️" : "🌙"}
              </span>
              {theme === "dark" ? t("account.lightMode") : t("account.darkMode")}
            </button>

            {/* Language switcher */}
            <div className="mt-1 flex items-center gap-1 px-3 py-1.5">
              <span
                aria-hidden="true"
                className="w-[18px] text-center text-[15px] text-faint"
              >
                🌐
              </span>
              <div
                role="group"
                aria-label={t("header.language")}
                className="flex flex-1 gap-1"
              >
                {LOCALES.map((l) => (
                  <button
                    key={l}
                    role="menuitemradio"
                    aria-checked={l === locale}
                    onClick={() => setLocale(l)}
                    className="flex-1 cursor-pointer rounded-md border py-1 text-[12.5px] font-semibold transition-colors"
                    style={
                      l === locale
                        ? {
                            borderColor: "var(--accent-ink)",
                            color: "var(--accent-ink)",
                            background:
                              "color-mix(in srgb, var(--accent) 10%, transparent)",
                          }
                        : {
                            borderColor: "var(--border-2)",
                            color: "var(--c5b)",
                          }
                    }
                  >
                    {LOCALE_LABELS[l]}
                  </button>
                ))}
              </div>
            </div>

            {loggedIn && (
              <>
                <div className="mx-2 my-1.5 h-px bg-border" />
                <button
                  role="menuitem"
                  onClick={async () => {
                    setOpen(false);
                    await logout();
                    track("logout", { method: "password" });
                    router.push("/");
                  }}
                  className={`${item} font-medium text-danger`}
                >
                  {t("account.logout")}
                </button>
              </>
            )}
          </div>
        </>
      )}
    </div>
  );
}
