import { ClerkProvider } from "@clerk/nextjs";
import type { Metadata } from "next";
import { Hanken_Grotesk, JetBrains_Mono } from "next/font/google";
import "./globals.css";

// Feed the design's typefaces into the existing --font-geist-* CSS variables
// so Tailwind's font-sans / font-mono pick them up without touching globals.css.
const fontSans = Hanken_Grotesk({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const fontMono = JetBrains_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Shush",
  description: "Self-hosted secret manager",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`dark ${fontSans.variable} ${fontMono.variable} h-full antialiased`}
    >
      <body className="min-h-full">
        <ClerkProvider>{children}</ClerkProvider>
      </body>
    </html>
  );
}
