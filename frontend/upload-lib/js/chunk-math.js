/** @param {number} totalSize @param {number} chunkSize */
export function totalChunks(totalSize, chunkSize) {
  if (totalSize <= 0 || chunkSize <= 0) {
    throw new Error(
      `invalid size parameters: total_size=${totalSize} chunk_size=${chunkSize}`,
    );
  }
  return Math.ceil(totalSize / chunkSize);
}

/** @param {number} totalSize @param {number} chunkSize @param {number} chunkNumber @param {number} totalChunksCount */
export function expectedChunkSize(totalSize, chunkSize, chunkNumber, totalChunksCount) {
  if (chunkNumber === totalChunksCount - 1) {
    return totalSize - chunkNumber * chunkSize;
  }
  return chunkSize;
}
