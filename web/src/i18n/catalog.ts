// Assembles the per-namespace message modules into a flat English catalog.
// Keys are "<namespace>.<key>". i18n.test.mjs checks placeholder/tag parity
// against the interpolation fields each message declares.

import * as providers from "./messages/providers.ts";
import * as app from "./messages/app.ts";
import * as interfaceMessages from "./messages/interface.ts";
import * as english from "./messages/english.ts";
import * as settings from "./messages/settings.ts";
import * as common from "./messages/common.ts";

const namespaces = {
  providers,
  settings,
  english,
  interface: interfaceMessages,
  app,
  common,
};

type Namespaces = typeof namespaces;
type NamespaceName = keyof Namespaces & string;

export type MessageKey = {
  [N in NamespaceName]: `${N}.${keyof Namespaces[N]["en"] & string}`;
}[NamespaceName];

export type Catalog = Record<MessageKey, string>;

function flatten(locale: "en"): Catalog {
  const out: Record<string, string> = {};
  for (const [ns, mod] of Object.entries(namespaces)) {
    for (const [key, value] of Object.entries(mod[locale] as Record<string, string>)) out[`${ns}.${key}`] = value;
  }
  return out as Catalog;
}

export const catalogs = { en: flatten("en") } as const;

export const namespaceNames = Object.keys(namespaces);
