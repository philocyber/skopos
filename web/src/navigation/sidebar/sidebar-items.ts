import {
  Activity,
  Ban,
  BellRing,
  BookOpen,
  Bot,
  Brain,
  Bug,
  ClipboardList,
  FolderOpen,
  FolderSync,
  LayoutDashboard,
  type LucideIcon,
  MessageSquare,
  Network,
  Plug,
  Radio,
  ScrollText,
  Settings2,
  ShieldAlert,
  Sparkles,
  Target,
  Terminal,
  Wrench,
} from "lucide-react";

import { translate as swt } from "@/i18n/runtime";

export type NavBadge = "new" | "soon";

export interface NavSubItem {
  id: string;
  title: string;
  url: string;
  icon?: LucideIcon;
  badge?: NavBadge;
  disabled?: boolean;
  newTab?: boolean;
}

interface NavItemBase {
  id: string;
  title: string;
  icon?: LucideIcon;
  badge?: NavBadge;
  disabled?: boolean;
  newTab?: boolean;
}

export interface NavMainLinkItem extends NavItemBase {
  url: string;
  subItems?: never;
}

export interface NavMainParentItem extends NavItemBase {
  subItems: NavSubItem[];
}

export type NavMainItem = NavMainLinkItem | NavMainParentItem;

export interface NavGroup {
  id: number;
  label?: string;
  // Pinned groups render below the scrolling list so they stay reachable on short screens.
  pinned?: boolean;
  items: NavMainItem[];
}

const link = (id: string, title: () => string, url: string, icon: LucideIcon): NavMainLinkItem => ({
  id,
  get title() {
    return title();
  },
  url,
  icon,
});

export const sidebarItems: NavGroup[] = [
  {
    id: 1,
    get label() {
      return swt("english.e197");
    },
    items: [
      link("dashboard", () => swt("interface.m2871"), "/dashboard", LayoutDashboard),
      link("chat", () => swt("interface.m2872"), "/chat", MessageSquare),
      link("tasks", () => swt("interface.m0190"), "/function/tasks", Target),
      link("findings", () => swt("interface.m0188"), "/function/findings", Bug),
      link("assets", () => swt("interface.m0222"), "/function/assets", Network),
      link("sync", () => swt("interface.m0445"), "/function/sync", FolderSync),
      link("workspace", () => swt("interface.m1113"), "/function/workspace", FolderOpen),
    ],
  },
  {
    id: 2,
    get label() {
      return swt("english.e198");
    },
    items: [
      link("traffic", () => swt("interface.m1067"), "/function/traffic", Activity),
      link("commands", () => swt("interface.m0287"), "/function/commands", Terminal),
      link("llm-records", () => swt("interface.m0426"), "/function/llm-records", Radio),
      link("logs", () => swt("interface.m2873"), "/system/logs", ScrollText),
    ],
  },
  {
    id: 3,
    get label() {
      return swt("english.e199");
    },
    items: [
      link("llm", () => "LLM", "/system/llm", Brain),
      link("agents", () => swt("english.e194"), "/system/agents", Bot),
      link("mcp", () => "MCP", "/system/mcp", Plug),
      link("skills", () => swt("english.e195"), "/system/skills", Sparkles),
      link("tools", () => swt("english.e196"), "/system/tools", Wrench),
      link("notify", () => swt("interface.m1541"), "/system/notify", BellRing),
      link("settings", () => swt("interface.m1644"), "/system/settings", Settings2),
    ],
  },
  {
    id: 4,
    get label() {
      return swt("english.e200");
    },
    items: [
      link("intercept", () => swt("interface.m1239"), "/system/intercept", ShieldAlert),
      link("asset-intercept", () => swt("interface.m1166"), "/system/intercept/assets", Ban),
      link("approvals", () => swt("interface.m2145"), "/system/intercept/approvals", ClipboardList),
    ],
  },
  {
    id: 5,
    pinned: true,
    get label() {
      return swt("english.e192");
    },
    items: [link("docs", () => swt("english.e193"), "/docs", BookOpen)],
  },
];

// Resolves the sidebar entry (and its group) that owns a pathname; the longest matching url wins so
// /system/intercept/approvals is not claimed by /system/intercept.
export function findNavEntry(pathname: string): { group: NavGroup; item: NavMainLinkItem | NavSubItem } | undefined {
  const route = pathname.replace(/\/+$/, "");
  let best: { group: NavGroup; item: NavMainLinkItem | NavSubItem } | undefined;
  for (const group of sidebarItems) {
    for (const main of group.items) {
      const entries: (NavMainLinkItem | NavSubItem)[] = main.subItems ?? [main as NavMainLinkItem];
      for (const item of entries) {
        if (!item.url || !(route === item.url || route.startsWith(`${item.url}/`))) continue;
        if (!best || item.url.length > best.item.url.length) best = { group, item };
      }
    }
  }
  return best;
}
