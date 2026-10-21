const TOKEN_KEY = "skopos_token";
const ROUTE_COOKIE = "skopos_session";
const COOKIE_MAX_AGE = 7 * 24 * 60 * 60; // Seven days in seconds.

export interface CurrentUser {
  id: string;
  name: string;
  username: string;
  email: string;
  avatar: string;
  role: string;
}

export const auth = {
  getToken(): string | null {
    if (typeof window === "undefined") return null;
    // Remove credentials left in cookies by older versions. The API uses the
    // origin-scoped localStorage token through explicit request credentials.
    if (document.cookie.split(";").some((entry) => entry.trim().startsWith(`${TOKEN_KEY}=`))) {
      document.cookie = `${TOKEN_KEY}=; path=/; max-age=0`;
    }
    // Mock mode returns a fake token without real authentication so route guards allow the demo.
    return localStorage.getItem(TOKEN_KEY) ?? (process.env.NEXT_PUBLIC_MOCK === "1" ? "mock-demo" : null);
  },

  setToken(token: string): void {
    localStorage.setItem(TOKEN_KEY, token);
    // This non-secret hint only controls page redirects. It grants no API access.
    document.cookie = `${TOKEN_KEY}=; path=/; max-age=0`;
    document.cookie = `${ROUTE_COOKIE}=1; path=/; max-age=${COOKIE_MAX_AGE}; SameSite=Lax`;
  },

  clearToken(): void {
    localStorage.removeItem(TOKEN_KEY);
    document.cookie = `${TOKEN_KEY}=; path=/; max-age=0`;
    document.cookie = `${ROUTE_COOKIE}=; path=/; max-age=0`;
  },

  // Read JWT sub for display only; this does not verify signatures.
  getCurrentUser(): CurrentUser | null {
    const token = this.getToken();
    if (!token) return null;
    try {
      const parts = token.split(".");
      if (parts.length !== 3) return null;
      // base64url → base64
      const payload = JSON.parse(atob(parts[1].replace(/-/g, "+").replace(/_/g, "/")));
      const username: string = payload.sub ?? "SKOPOS";
      return { id: "1", name: username, username, email: "", avatar: "", role: "operator" };
    } catch {
      return null;
    }
  },
};
