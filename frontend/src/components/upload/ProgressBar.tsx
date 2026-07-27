import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import type { FileUploadState } from "@/hooks/useUploadStore";

function statusVariant(status: FileUploadState["status"]) {
  if (status === "complete") return "success" as const;
  if (status === "error") return "error" as const;
  if (status === "uploading") return "default" as const;
  return "muted" as const;
}

export function ProgressBar({ files }: { files: FileUploadState[] }) {
  if (files.length === 0) return null;
  return (
    <ul className="mt-4 space-y-3">
      {files.map((file) => (
        <li key={file.id} className="rounded-lg border border-zinc-200 p-3">
          <div className="mb-2 flex items-center justify-between gap-2">
            <div className="min-w-0 truncate text-sm font-medium">{file.name}</div>
            <Badge variant={statusVariant(file.status)}>{file.status}</Badge>
          </div>
          <Progress value={file.percentage} />
          <div className="mt-1 flex justify-between text-xs text-zinc-500">
            <span>{(file.size / 1024).toFixed(1)} KB</span>
            <span>{file.percentage.toFixed(0)}%</span>
          </div>
          {file.error ? <p className="mt-1 text-xs text-red-600">{file.error}</p> : null}
        </li>
      ))}
    </ul>
  );
}
