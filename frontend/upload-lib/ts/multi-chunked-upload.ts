import { ChunkedUploader, type ChunkedUploaderOptions } from "./chunked-upload";
import type { UploadProgress } from "./types";

export type MultiFileProgress = {
  fileName: string;
  progress: UploadProgress;
  status: "pending" | "uploading" | "complete" | "error";
  error?: string;
};

export type MultiChunkedOptions = ChunkedUploaderOptions & {
  fileConcurrency?: number;
  chunkConcurrency?: number;
  onFileProgress?: (update: MultiFileProgress) => void;
};

/** Upload multiple files with bounded per-file chunked sessions. */
export async function multiChunkedUpload(
  files: File[],
  options: MultiChunkedOptions = {},
): Promise<MultiFileProgress[]> {
  const fileConcurrency = options.fileConcurrency ?? 2;
  const chunkConcurrency = options.chunkConcurrency ?? 3;
  const results: MultiFileProgress[] = files.map((f) => ({
    fileName: f.name,
    progress: { uploaded: 0, total: 0, percentage: 0 },
    status: "pending",
  }));

  let index = 0;

  async function worker() {
    while (index < files.length) {
      const i = index++;
      const file = files[i];
      results[i].status = "uploading";
      options.onFileProgress?.(results[i]);
      try {
        const uploader = new ChunkedUploader(file, {
          ...options,
          onProgress: (progress) => {
            results[i].progress = progress;
            options.onFileProgress?.({ ...results[i] });
            options.onProgress?.(progress);
          },
        });
        await uploader.upload(chunkConcurrency);
        results[i].status = "complete";
        results[i].progress = {
          ...results[i].progress,
          percentage: 100,
        };
      } catch (error) {
        results[i].status = "error";
        results[i].error = error instanceof Error ? error.message : String(error);
      }
      options.onFileProgress?.({ ...results[i] });
    }
  }

  await Promise.all(
    Array.from({ length: Math.min(fileConcurrency, files.length) }, () => worker()),
  );
  return results;
}
