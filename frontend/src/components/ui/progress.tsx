import { cn } from "@/lib/utils";

export function Progress({
  value = 0,
  className,
}: {
  value?: number;
  className?: string;
}) {
  const clamped = Math.max(0, Math.min(100, value));
  return (
    <div
      className={cn("relative h-2 w-full overflow-hidden rounded-full bg-zinc-200", className)}
      role="progressbar"
      aria-valuenow={clamped}
      aria-valuemin={0}
      aria-valuemax={100}
    >
      <div
        className="h-full bg-emerald-600 transition-all duration-200"
        style={{ width: `${clamped}%` }}
      />
    </div>
  );
}
