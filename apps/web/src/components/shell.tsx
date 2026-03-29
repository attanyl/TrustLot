"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  { href: "/", label: "Dashboard" },
  { href: "/health", label: "Health" },
  { href: "/exceptions", label: "Exceptions" },
  { href: "/recon-runs", label: "Recon Runs" },
  { href: "/trust", label: "Trust Scores" },
  { href: "/lineage", label: "Lineage" },
  { href: "/replay", label: "Replay" },
];

export function Shell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  return (
    <div className="flex h-screen bg-zinc-900 text-zinc-100">
      <nav className="w-56 flex-shrink-0 bg-zinc-950 border-r border-zinc-800 flex flex-col">
        <div className="px-4 py-4 border-b border-zinc-800">
          <h1 className="text-sm font-bold tracking-widest text-zinc-400 uppercase">
            TrustLot
          </h1>
          <p className="text-[10px] text-zinc-600 mt-1">operator console</p>
        </div>
        <div className="flex-1 py-2">
          {navItems.map((item) => {
            const isActive =
              item.href === "/"
                ? pathname === "/"
                : pathname.startsWith(item.href);
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`block px-4 py-2 text-sm ${
                  isActive
                    ? "bg-zinc-800 text-zinc-100 border-l-2 border-zinc-400"
                    : "text-zinc-500 hover:text-zinc-300 hover:bg-zinc-900 border-l-2 border-transparent"
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </div>
        <div className="px-4 py-3 border-t border-zinc-800 text-[10px] text-zinc-700">
          synthetic data only
        </div>
      </nav>
      <main className="flex-1 overflow-auto p-6">{children}</main>
    </div>
  );
}
