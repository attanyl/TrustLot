import type { Metadata } from "next";
import { Shell } from "@/components/shell";
import "./globals.css";

export const metadata: Metadata = {
  title: "TrustLot",
  description: "Explainable wealth data trust layer — operator console",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="bg-zinc-900 text-zinc-100 antialiased">
        <Shell>{children}</Shell>
      </body>
    </html>
  );
}
