// Self-hosted typefaces keep the embedded console independent of font CDNs.
import localFont from "next/font/local";

const inter = localFont({
  src: [
    { path: "./brand/Inter-Regular.ttf", weight: "400" },
    { path: "./brand/Inter-Medium.ttf", weight: "500" },
    { path: "./brand/Inter-SemiBold.ttf", weight: "600" },
  ],
  variable: "--font-inter",
  display: "swap",
});
const jetBrainsMono = localFont({
  src: "./brand/JetBrainsMono-Regular.ttf",
  weight: "400",
  variable: "--font-jetbrains-mono",
  display: "swap",
});

export const fontVars = `${inter.variable} ${jetBrainsMono.variable}`;
