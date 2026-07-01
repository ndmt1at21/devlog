"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useRouter } from "next/navigation";
import { LOCALE_COOKIE, makeT, type Locale, type TFunc } from "./dictionaries";

interface LocaleContextValue {
  locale: Locale;
  setLocale: (l: Locale) => void;
  t: TFunc;
}

const LocaleContext = createContext<LocaleContextValue | null>(null);

export function LocaleProvider({
  initialLocale,
  children,
}: {
  initialLocale: Locale;
  children: ReactNode;
}) {
  const [locale, setLocaleState] = useState<Locale>(initialLocale);
  const router = useRouter();

  const setLocale = useCallback(
    (l: Locale) => {
      document.cookie = `${LOCALE_COOKIE}=${l}; path=/; max-age=31536000; samesite=lax`;
      setLocaleState(l);
      // Re-render Server Components (metadata, not-found) with the new locale.
      router.refresh();
    },
    [router],
  );

  const t = useMemo(() => makeT(locale), [locale]);
  const value = useMemo<LocaleContextValue>(
    () => ({ locale, setLocale, t }),
    [locale, setLocale, t],
  );

  return (
    <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>
  );
}

export function useI18n(): LocaleContextValue {
  const ctx = useContext(LocaleContext);
  if (!ctx) throw new Error("useI18n must be used within <LocaleProvider>");
  return ctx;
}

/** Shorthand for the translator. */
export function useT(): TFunc {
  return useI18n().t;
}
