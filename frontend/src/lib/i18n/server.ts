// Server-only locale helpers: read the locale cookie in Server Components and
// build a translator for SSR text (metadata, not-found, etc.).
import "server-only";
import { cookies } from "next/headers";
import {
  LOCALE_COOKIE,
  makeT,
  type Locale,
  type TFunc,
} from "./dictionaries";

export async function getLocale(): Promise<Locale> {
  const value = (await cookies()).get(LOCALE_COOKIE)?.value;
  return value === "en" ? "en" : "vi";
}

export async function getT(): Promise<TFunc> {
  return makeT(await getLocale());
}
