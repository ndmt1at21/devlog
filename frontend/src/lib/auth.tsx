"use client";

// Client auth/session context. Hydrates from an SSR-provided value (so the
// header renders correctly on first paint) then re-validates against
// /api/auth/me. Premium status drives ad-hiding and series unlocking.

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { api } from "./api";
import type { MeResponse, SessionUser } from "./types";

interface AuthContextValue {
  user: SessionUser | null;
  loading: boolean;
  premium: boolean;
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({
  initial,
  children,
}: {
  initial: MeResponse;
  children: ReactNode;
}) {
  const [user, setUser] = useState<SessionUser | null>(
    initial.authenticated ? (initial.user ?? null) : null,
  );
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const me = await api.me();
      setUser(me.authenticated ? (me.user ?? null) : null);
    } catch {
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await api.logout();
    } finally {
      setUser(null);
    }
  }, []);

  // Re-validate once on mount in case the SSR snapshot is stale. This is a
  // fetch-on-mount from an external system (the /api/auth/me endpoint), not a
  // synchronous cascading render.
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- fetch on mount
    void refresh();
  }, [refresh]);

  const value = useMemo<AuthContextValue>(
    () => ({ user, loading, premium: user?.premium ?? false, refresh, logout }),
    [user, loading, refresh, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within <AuthProvider>");
  return ctx;
}
