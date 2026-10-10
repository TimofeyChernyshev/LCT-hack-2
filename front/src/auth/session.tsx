import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { api, clearTokens, hasToken, saveTokens } from "../api/client";
import type { Role, TokenPair, User } from "../api/types";

type AuthState = {
  user: User | null;
  ready: boolean;
  login: (email: string, password: string, expectedRole?: Role) => Promise<User>;
  logout: () => Promise<void>;
};

export function roleMismatchMessage(actual: Role): string {
  if (actual === "employer") {
    return "Этот аккаунт зарегистрирован как работодатель. Войдите через кабинет работодателя.";
  }
  if (actual === "candidate") {
    return "Этот аккаунт зарегистрирован как соискатель. Войдите через кабинет соискателя.";
  }
  return "Этот аккаунт нельзя использовать в выбранном кабинете.";
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    let live = true;
    if (!hasToken()) {
      setReady(true);
      return;
    }
    api<User>("auth", "/auth/me")
      .then((me) => {
        if (live) setUser(me);
      })
      .catch(() => {
        clearTokens();
      })
      .finally(() => {
        if (live) setReady(true);
      });
    return () => {
      live = false;
    };
  }, []);

  const value = useMemo<AuthState>(
    () => ({
      user,
      ready,
      async login(email, password, expectedRole) {
        const pair = await api<TokenPair>("auth", "/auth/login", {
          method: "POST",
          auth: false,
          body: JSON.stringify({ email, password }),
        });
        if (expectedRole && pair.user.role !== expectedRole) {
          try {
            await api("auth", "/auth/logout", {
              method: "POST",
              auth: false,
              headers: { Authorization: `Bearer ${pair.accessToken}` },
              body: JSON.stringify({ refreshToken: pair.refreshToken }),
            });
          } catch {
            // Клиент сессию не сохраняет, даже если отзыв токена не прошёл.
          }
          throw new Error(roleMismatchMessage(pair.user.role));
        }
        saveTokens(pair);
        setUser(pair.user);
        return pair.user;
      },
      async logout() {
        const refreshToken = localStorage.getItem("fsp_refresh");
        try {
          if (refreshToken) {
            await api("auth", "/auth/logout", {
              method: "POST",
              body: JSON.stringify({ refreshToken }),
            });
          }
        } finally {
          clearTokens();
          setUser(null);
        }
      },
    }),
    [user, ready],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("AuthProvider отсутствует");
  return value;
}
