import { ChunkedUploader } from "./chunked-upload.js";

/**
 * @param {File[]} files
 * @param {{ fileConcurrency?: number, chunkConcurrency?: number, onFileProgress?: Function, onProgress?: Function, baseUrl?: string, chunkSize?: number, signal?: AbortSignal }} [options]
 */
export async function multiChunkedUpload(files, options = {}) {
  const fileConcurrency = options.fileConcurrency ?? 2;
  const chunkConcurrency = options.chunkConcurrency ?? 3;
  const results = files.map((f) => ({
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
        results[i].progress = { ...results[i].progress, percentage: 100 };
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
