/** Ceiling division for chunk count — mirrors Go chunked.TotalChunks */
export function totalChunks(totalSize: number, chunkSize: number): number {
  if (totalSize <= 0 || chunkSize <= 0) {
    throw new Error(
      `invalid size parameters: total_size=${totalSize} chunk_size=${chunkSize}`,
    );
  }
  return Math.ceil(totalSize / chunkSize);
}

/** Expected byte length of a 0-based chunk — mirrors Go chunked.ExpectedChunkSize */
export function expectedChunkSize(
  totalSize: number,
  chunkSize: number,
  chunkNumber: number,
  totalChunksCount: number,
): number {
  if (chunkNumber === totalChunksCount - 1) {
    return totalSize - chunkNumber * chunkSize;
  }
  return chunkSize;
}
