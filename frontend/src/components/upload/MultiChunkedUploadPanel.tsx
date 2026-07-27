import { useMutation } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { multiChunkedUpload } from "@upload-lib/multi-chunked-upload";
import { Button } from "@/components/ui/button";
import { FileDropzone } from "@/components/upload/FileDropzone";
import { ProgressBar } from "@/components/upload/ProgressBar";
import { useUploadStore, type FileUploadState } from "@/hooks/useUploadStore";
import { API_BASE } from "@/lib/api";

export function MultiChunkedUploadPanel() {
  const [selected, setSelected] = useState<File[]>([]);
  const abortRef = useRef<AbortController | null>(null);
  const stateRef = useRef<FileUploadState[]>([]);
  const { files, setFiles, setMessage, reset } = useUploadStore();

  const mutation = useMutation({
    mutationFn: async (uploadFiles: File[]) => {
      const controller = new AbortController();
      abortRef.current = controller;
      stateRef.current = uploadFiles.map((f, i) => ({
        id: `${f.name}-${i}`,
        name: f.name,
        size: f.size,
        percentage: 0,
        status: "uploading" as const,
      }));
      setFiles(stateRef.current);

      return multiChunkedUpload(uploadFiles, {
        baseUrl: API_BASE,
        signal: controller.signal,
        fileConcurrency: 2,
        chunkConcurrency: 3,
        onFileProgress: (update) => {
          stateRef.current = stateRef.current.map((f) => {
            if (f.name !== update.fileName) return f;
            return {
              ...f,
              percentage: update.progress.percentage,
              status:
                update.status === "pending"
                  ? "idle"
                  : update.status === "uploading"
                    ? "uploading"
                    : update.status === "complete"
                      ? "complete"
                      : "error",
              error: update.error,
            };
          });
          setFiles([...stateRef.current]);
        },
      });
    },
    onSuccess: (results) => {
      const next = results.map((r, i) => ({
        id: `${r.fileName}-${i}`,
        name: r.fileName,
        size: selected[i]?.size ?? 0,
        percentage: r.progress.percentage,
        status:
          r.status === "complete"
            ? ("complete" as const)
            : r.status === "error"
              ? ("error" as const)
              : ("idle" as const),
        error: r.error,
      }));
      stateRef.current = next;
      setFiles(next);
      const ok = results.filter((r) => r.status === "complete").length;
      setMessage(`Uploaded ${ok}/${results.length} file(s) via chunked multi-file`);
    },
    onError: (error: Error) => {
      setMessage(error.message);
    },
  });

  return (
    <div className="space-y-4">
      <FileDropzone
        multiple
        disabled={mutation.isPending}
        onFiles={(next) => {
          reset();
          setSelected(next);
          const initial = next.map((f, i) => ({
            id: `${f.name}-${i}`,
            name: f.name,
            size: f.size,
            percentage: 0,
            status: "idle" as const,
          }));
          stateRef.current = initial;
          setFiles(initial);
        }}
      />
      <div className="flex gap-2">
        <Button
          disabled={selected.length === 0 || mutation.isPending}
          onClick={() => mutation.mutate(selected)}
        >
          {mutation.isPending ? "Uploading…" : "Upload (chunked multi)"}
        </Button>
        <Button
          variant="outline"
          disabled={!mutation.isPending}
          onClick={() => abortRef.current?.abort()}
        >
          Cancel
        </Button>
      </div>
      <ProgressBar files={files} />
    </div>
  );
}
