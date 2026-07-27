import { useMutation } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { ChunkedUploader } from "@upload-lib/chunked-upload";
import { Button } from "@/components/ui/button";
import { FileDropzone } from "@/components/upload/FileDropzone";
import { ProgressBar } from "@/components/upload/ProgressBar";
import { useUploadStore } from "@/hooks/useUploadStore";
import { API_BASE } from "@/lib/api";

export function ChunkedUploadPanel() {
  const [selected, setSelected] = useState<File | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const { files, setFiles, setMessage, reset } = useUploadStore();

  const mutation = useMutation({
    mutationFn: async (file: File) => {
      const controller = new AbortController();
      abortRef.current = controller;
      setFiles([
        {
          id: file.name,
          name: file.name,
          size: file.size,
          percentage: 0,
          status: "uploading",
        },
      ]);

      const uploader = new ChunkedUploader(file, {
        baseUrl: API_BASE,
        signal: controller.signal,
        onProgress: (p) => {
          setFiles([
            {
              id: file.name,
              name: file.name,
              size: file.size,
              percentage: p.percentage,
              status: "uploading",
            },
          ]);
        },
      });
      return uploader.upload(3);
    },
    onSuccess: (result) => {
      setFiles([
        {
          id: result.filename,
          name: result.filename,
          size: result.size,
          percentage: 100,
          status: "complete",
        },
      ]);
      setMessage(`Chunked upload complete: ${result.filename}`);
    },
    onError: (error: Error) => {
      if (selected) {
        setFiles([
          {
            id: selected.name,
            name: selected.name,
            size: selected.size,
            percentage: 0,
            status: "error",
            error: error.message,
          },
        ]);
      }
      setMessage(error.message);
    },
  });

  return (
    <div className="space-y-4">
      <FileDropzone
        disabled={mutation.isPending}
        onFiles={(next) => {
          reset();
          const file = next[0] ?? null;
          setSelected(file);
          if (file) {
            setFiles([
              {
                id: file.name,
                name: file.name,
                size: file.size,
                percentage: 0,
                status: "idle",
              },
            ]);
          }
        }}
      />
      <div className="flex gap-2">
        <Button
          disabled={!selected || mutation.isPending}
          onClick={() => selected && mutation.mutate(selected)}
        >
          {mutation.isPending ? "Uploading…" : "Upload (chunked)"}
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
