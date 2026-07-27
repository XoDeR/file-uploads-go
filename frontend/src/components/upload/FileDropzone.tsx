import { useRef, useState } from "react";
import { Upload } from "lucide-react";
import { cn } from "@/lib/utils";

type Props = {
  multiple?: boolean;
  disabled?: boolean;
  onFiles: (files: File[]) => void;
};

export function FileDropzone({ multiple = false, disabled, onFiles }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);

  function handleFiles(list: FileList | null) {
    if (!list || list.length === 0) return;
    const files = Array.from(list);
    onFiles(multiple ? files : files.slice(0, 1));
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") inputRef.current?.click();
      }}
      onClick={() => !disabled && inputRef.current?.click()}
      onDragOver={(e) => {
        e.preventDefault();
        if (!disabled) setDragging(true);
      }}
      onDragLeave={() => setDragging(false)}
      onDrop={(e) => {
        e.preventDefault();
        setDragging(false);
        if (!disabled) handleFiles(e.dataTransfer.files);
      }}
      className={cn(
        "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed border-zinc-300 bg-zinc-50 px-6 py-10 text-center transition-colors",
        dragging && "border-emerald-500 bg-emerald-50",
        disabled && "pointer-events-none opacity-50",
      )}
    >
      <Upload className="h-8 w-8 text-zinc-400" />
      <div className="text-sm font-medium text-zinc-800">
        Drop {multiple ? "files" : "a file"} here or click to browse
      </div>
      <div className="text-xs text-zinc-500">Stored on local disk via the Go backend</div>
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        multiple={multiple}
        disabled={disabled}
        onChange={(e) => handleFiles(e.target.files)}
      />
    </div>
  );
}
