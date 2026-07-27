export type UploadProgress = {
  uploaded: number;
  total: number;
  percentage: number;
  bytesUploaded?: number;
  bytesTotal?: number;
};

export type ChunkedSession = {
  id: string;
  filename: string;
  total_size: number;
  chunk_size: number;
  total_chunks: number;
};

export type ChunkedStatus = {
  upload_id: string;
  filename: string;
  uploaded_chunks: number;
  total_chunks: number;
  missing_chunks: number[];
  expires_at: string;
};

export type StreamUploadResult = {
  status: string;
  files: Array<{ filename: string; upload_id: string }>;
  upload_id: string;
  message: string;
};

export type ProgressEvent = {
  upload_id: string;
  filename: string;
  bytes_read: number;
  total_bytes: number;
  percentage: number;
  status: string;
};
