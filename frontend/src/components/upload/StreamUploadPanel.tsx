import { useMutation } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { streamUpload } from "@upload-lib/stream-upload";
import { Button } from "@/components/ui/button";
import { FileDropzone } from "@/components/upload/FileDropzone";
import { ProgressBar } from "@/components/upload/ProgressBar";
import { useUploadStore } from "@/hooks/useUploadStore";
import { API_BASE } from "@/lib/api";

type Props = { multiple?: boolean };

export function StreamUploadPanel({ multiple = false }: Props) {
  const [selected, setSelected] = useState<File[]>([]);
  const abortRef = useRef<AbortController | null>(null);
  const { files, setFiles, setMessage, reset } = useUploadStore();

  const mutation = useMutation({
    mutationFn: async (uploadFiles: File[]) => {
      const controller = new AbortController();
      abortRef.current = controller;
      setFiles(
        uploadFiles.map((f, i) => ({
          id: `${f.name}-${i}`,
          name: f.name,
          size: f.size,
          percentage: 0,
          status: "uploading" as const,
        })),
      );

      return streamUpload(uploadFiles, {
        baseUrl: API_BASE,
        signal: controller.signal,
        onProgress: (p) => {
          setFiles(
            uploadFiles.map((f, i) => ({
              id: `${f.name}-${i}`,
              name: f.name,
              size: f.size,
              percentage: p.percentage,
              status: "uploading" as const,
            })),
          );
        },
      });
    },
    onSuccess: (result) => {
      setFiles(
        result.files.map((f, i) => ({
          id: `${f.filename}-${i}`,
          name: f.filename,
          size: selected[i]?.size ?? 0,
          percentage: 100,
          status: "complete" as const,
        })),
      );
      setMessage(result.message);
    },
    onError: (error: Error) => {
      setFiles(
        selected.map((f, i) => ({
          id: `${f.name}-${i}`,
          name: f.name,
          size: f.size,
          percentage: 0,
          status: "error" as const,
          error: error.message,
        })),
      );
      setMessage(error.message);
    },
  });

  return (
    <div className="space-y-4">
      <FileDropzone
        multiple={multiple}
        disabled={mutation.isPending}
        onFiles={(next) => {
          reset();
          setSelected(next);
          setFiles(
            next.map((f, i) => ({
              id: `${f.name}-${i}`,
              name: f.name,
              size: f.size,
              percentage: 0,
              status: "idle",
            })),
          );
        }}
      />
      <div className="flex gap-2">
        <Button
          disabled={selected.length === 0 || mutation.isPending}
          onClick={() => mutation.mutate(selected)}
        >
          {mutation.isPending ? "Uploading…" : "Upload (stream)"}
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
