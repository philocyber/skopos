"use client";

import { translate as swt } from "@/i18n/runtime";
import { useI18n } from "@/i18n";
import { type ReactNode, useEffect, useState } from "react";

import { usePathname } from "next/navigation";

import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { APP_CONFIG } from "@/config/app-config";
import { useCurrentUser } from "@/hooks/use-current-user";
import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import { sidebarItems } from "@/navigation/sidebar/sidebar-items";

import { AccountSwitcher } from "./sidebar/account-switcher";
import { LayoutControls } from "./sidebar/layout-controls";
import { SearchDialog } from "./sidebar/search-dialog";
import { ThemeSwitcher } from "./sidebar/theme-switcher";
import { UpdateBadge } from "./update-badge";

// Task details supply their own header, tabs, and padding; do not add the global header or padding.
function isFullBleed(pathname: string) {
  const p = (() => {
    try {
      return decodeURIComponent(pathname);
    } catch {
      return pathname;
    }
  })();
  // Static export uses trailingSlash, so the task-list pathname is /function/tasks/.
  // Remove the trailing slash before prefix checks or the list is mistaken for details and loses its header.
  const normalized = p.replace(/\/+$/, "");
  return normalized.startsWith("/function/tasks/") || normalized.startsWith("/function/findings/");
}

export function MainContent({ children }: { children: ReactNode }) {
  "use no memo";
  const { t: swt, locale: swLocale } = useI18n();

  const currentUser = useCurrentUser();
  const pathname = usePathname();
  const route = pathname.replace(/\/+$/, "");
  const routeTitle = sidebarItems
    .flatMap((group) => group.items.flatMap((item) => item.subItems ?? [item]))
    .filter((item) => item.url && (route === item.url || route.startsWith(`${item.url}/`)))
    .sort((a, b) => (b.url?.length ?? 0) - (a.url?.length ?? 0))[0]?.title;
  const [version, setVersion] = useState("");

  useEffect(() => {
    api
      .health()
      .then((h) => setVersion((h.version ?? "").replace(/^v(?=\d)/, "")))
      .catch(() => setVersion(""));
  }, []);

  if (isFullBleed(pathname)) {
    return <>{children}</>;
  }

  return (
    <>
      <header
        data-console="toolbar"
        className={cn(
          "flex h-14 shrink-0 items-center gap-2 border-b bg-card transition-[width,height] ease-linear",
          "[html[data-navbar-style=sticky]_&]:sticky [html[data-navbar-style=sticky]_&]:top-0 [html[data-navbar-style=sticky]_&]:z-40 [html[data-navbar-style=sticky]_&]:rounded-t-[inherit]",
        )}
      >
        <div className="flex w-full items-center justify-between px-4 lg:px-6">
          <div className="flex items-center gap-1 lg:gap-2">
            <SidebarTrigger className="-ml-1" />
            <Separator
              orientation="vertical"
              className="mx-2 data-[orientation=vertical]:h-4 data-[orientation=vertical]:self-center"
            />
            <span className="truncate text-sm font-medium">{routeTitle ?? APP_CONFIG.name}</span>
          </div>
          <div className="flex items-center gap-2">
            <SearchDialog />
            {version && (
              <span className="hidden font-mono text-[11px] text-muted-foreground tabular-nums lg:inline">{swt("interface.m0052")}{version}</span>
            )}
            <UpdateBadge />
            <LayoutControls />
            <ThemeSwitcher />
            <AccountSwitcher users={[currentUser]} />
          </div>
        </div>
      </header>
      <div data-console="workspace" className="min-h-0 min-w-0 flex-1 overflow-x-hidden p-4 has-data-[content-padding=false]:p-0 md:p-6 md:has-data-[content-padding=false]:p-0">
        {children}
      </div>
    </>
  );
}
