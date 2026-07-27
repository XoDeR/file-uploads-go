import { totalChunks } from "./chunk-math.js";

export class ChunkedUploader {
  /**
   * @param {File} file
   * @param {{ chunkSize?: number, baseUrl?: string, onProgress?: Function, signal?: AbortSignal }} [options]
   */
  constructor(file, options = {}) {
    this.file = file;
    this.chunkSize = options.chunkSize ?? 5 * 1024 * 1024;
    this.baseUrl = options.baseUrl ?? "/api/upload";
    this.uploadId = null;
    this.totalChunksCount = 0;
    this.uploadedChunks = new Set();
    this.onProgress = options.onProgress ?? (() => {});
    this.signal = options.signal;
  }

  async init() {
    this.throwIfAborted();
    const response = await fetch(`${this.baseUrl}/init`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        filename: this.file.name,
        total_size: this.file.size,
        chunk_size: this.chunkSize,
      }),
      signal: this.signal,
    });
    if (!response.ok) throw new Error(await response.text());
    const data = await response.json();
    this.uploadId = data.id;
    this.totalChunksCount = data.total_chunks;
    return data;
  }

  async uploadChunk(chunkNumber, retries = 3) {
    if (!this.uploadId) throw new Error("Call init() first");
    const start = chunkNumber * this.chunkSize;
    const end = Math.min(start + this.chunkSize, this.file.size);
    const chunk = this.file.slice(start, end);

    for (let attempt = 0; attempt < retries; attempt++) {
      this.throwIfAborted();
      try {
        const response = await fetch(
          `${this.baseUrl}/chunk?upload_id=${this.uploadId}&chunk=${chunkNumber}`,
          { method: "POST", body: chunk, signal: this.signal },
        );
        if (response.ok) {
          this.uploadedChunks.add(chunkNumber);
          this.onProgress({
            uploaded: this.uploadedChunks.size,
            total: this.totalChunksCount,
            percentage: (this.uploadedChunks.size / this.totalChunksCount) * 100,
          });
          return true;
        }
      } catch (error) {
        if (attempt === retries - 1) throw error;
        await new Promise((r) => setTimeout(r, 1000 * (attempt + 1)));
      }
    }
    return false;
  }

  async upload(concurrency = 3) {
    await this.init();
    if (this.totalChunksCount === 0) {
      this.totalChunksCount = totalChunks(this.file.size, this.chunkSize);
    }
    const chunks = Array.from({ length: this.totalChunksCount }, (_, i) => i);
    for (let i = 0; i < chunks.length; i += concurrency) {
      const batch = chunks.slice(i, i + concurrency);
      await Promise.all(batch.map((chunk) => this.uploadChunk(chunk)));
    }
    return this.complete();
  }

  async complete() {
    if (!this.uploadId) throw new Error("Call init() first");
    this.throwIfAborted();
    const response = await fetch(
      `${this.baseUrl}/complete?upload_id=${this.uploadId}`,
      { method: "POST", signal: this.signal },
    );
    if (!response.ok) throw new Error(await response.text());
    return response.json();
  }

  async resume() {
    if (!this.uploadId) throw new Error("Call init() first");
    this.throwIfAborted();
    const response = await fetch(
      `${this.baseUrl}/status?upload_id=${this.uploadId}`,
      { signal: this.signal },
    );
    if (!response.ok) throw new Error(await response.text());
    const status = await response.json();
    for (const chunkNumber of status.missing_chunks) {
      await this.uploadChunk(chunkNumber);
    }
    return this.complete();
  }

  throwIfAborted() {
    if (this.signal?.aborted) throw new Error("Upload aborted");
  }
}
