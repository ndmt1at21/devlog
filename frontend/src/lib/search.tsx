"use client";

// Global search query shared between the Header search box and the Home grid.
// Typing anywhere routes to Home so the results are always visible.

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { usePathname, useRouter } from "next/navigation";

interface SearchContextValue {
  query: string;
  setQuery: (q: string) => void;
  clear: () => void;
}

const SearchContext = createContext<SearchContextValue | null>(null);

export function SearchProvider({ children }: { children: ReactNode }) {
  const [query, setQueryState] = useState("");
  const pathname = usePathname();
  const router = useRouter();

  const setQuery = useCallback(
    (q: string) => {
      setQueryState(q);
      if (q && pathname !== "/") router.push("/");
    },
    [pathname, router],
  );

  const clear = useCallback(() => setQueryState(""), []);

  const value = useMemo<SearchContextValue>(
    () => ({ query, setQuery, clear }),
    [query, setQuery, clear],
  );

  return (
    <SearchContext.Provider value={value}>{children}</SearchContext.Provider>
  );
}

export function useSearch(): SearchContextValue {
  const ctx = useContext(SearchContext);
  if (!ctx) throw new Error("useSearch must be used within <SearchProvider>");
  return ctx;
}
