/**
 * @param {File|File[]} files
 * @param {{ baseUrl?: string, uploadId?: string, onProgress?: Function, signal?: AbortSignal }} [options]
 */
export function streamUpload(files, options = {}) {
  const list = Array.isArray(files) ? files : [files];
  const baseUrl = options.baseUrl ?? "/api/upload";
  const form = new FormData();
  for (const file of list) {
    form.append("file", file, file.name);
  }
  const totalBytes = list.reduce((sum, f) => sum + f.size, 0);

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${baseUrl}/stream`);
    if (options.uploadId) {
      xhr.setRequestHeader("X-Upload-ID", options.uploadId);
    }
    xhr.upload.onprogress = (event) => {
      if (!options.onProgress) return;
      const loaded = event.lengthComputable ? event.loaded : 0;
      const total = event.lengthComputable ? event.total : totalBytes;
      options.onProgress({
        uploaded: 0,
        total: list.length,
        percentage: total > 0 ? (loaded / total) * 100 : 0,
        bytesUploaded: loaded,
        bytesTotal: total,
      });
    };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText));
        } catch {
          reject(new Error("Invalid JSON response"));
        }
      } else {
        reject(new Error(xhr.responseText || `Upload failed (${xhr.status})`));
      }
    };
    xhr.onerror = () => reject(new Error("Network error during upload"));
    xhr.onabort = () => reject(new Error("Upload aborted"));
    if (options.signal) {
      if (options.signal.aborted) {
        xhr.abort();
        return;
      }
      options.signal.addEventListener("abort", () => xhr.abort(), { once: true });
    }
    xhr.send(form);
  });
}
