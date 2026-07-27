import * as React from "react";
import { cn } from "@/lib/utils";

export function Badge({
  children,
  className,
  variant = "default",
}: {
  children: React.ReactNode;
  className?: string;
  variant?: "default" | "success" | "error" | "muted";
}) {
  const styles = {
    default: "bg-zinc-900 text-zinc-50",
    success: "bg-emerald-100 text-emerald-800",
    error: "bg-red-100 text-red-800",
    muted: "bg-zinc-100 text-zinc-700",
  }[variant];

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium",
        styles,
        className,
      )}
    >
      {children}
    </span>
  );
}
