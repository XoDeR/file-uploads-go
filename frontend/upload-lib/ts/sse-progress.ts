import type { ProgressEvent } from "./types";

export type SSEProgressOptions = {
  baseUrl?: string;
  onProgress: (event: ProgressEvent) => void;
  onError?: (error: Event) => void;
};

/** Subscribe to server-sent upload progress events. */
export function subscribeProgress(
  uploadId: string,
  options: SSEProgressOptions,
): () => void {
  const baseUrl = options.baseUrl ?? "/api/upload";
  const source = new EventSource(
    `${baseUrl}/progress?upload_id=${encodeURIComponent(uploadId)}`,
  );

  source.onmessage = (message) => {
    try {
      const data = JSON.parse(message.data) as ProgressEvent;
      options.onProgress(data);
      if (data.status === "completed" || data.status === "error") {
        source.close();
      }
    } catch {
      // ignore malformed frames
    }
  };

  source.onerror = (event) => {
    options.onError?.(event);
  };

  return () => source.close();
}
