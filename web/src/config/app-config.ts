import { translate as swt } from "@/i18n/runtime";
import packageJson from "../../package.json";

const currentYear = new Date().getFullYear();

export const APP_CONFIG = {
  name: "Skopos",
  version: packageJson.version,
  copyright: `© ${currentYear}, Skopos. Based on ARTEX by Autumn-27.`,
  meta: {
    get title() { return swt("interface.m2410"); },
    get description() { return swt("interface.m2411"); },
  },
};
