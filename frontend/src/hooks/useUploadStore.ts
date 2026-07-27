import { create } from "zustand";

export type FileUploadState = {
  id: string;
  name: string;
  size: number;
  percentage: number;
  status: "idle" | "uploading" | "complete" | "error";
  error?: string;
};

type UploadStore = {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  files: FileUploadState[];
  setFiles: (files: FileUploadState[]) => void;
  upsertFile: (file: FileUploadState) => void;
  reset: () => void;
  message: string | null;
  setMessage: (message: string | null) => void;
};

export const useUploadStore = create<UploadStore>((set) => ({
  activeTab: "stream",
  setActiveTab: (tab) => set({ activeTab: tab }),
  files: [],
  setFiles: (files) => set({ files }),
  upsertFile: (file) =>
    set((state) => {
      const idx = state.files.findIndex((f) => f.id === file.id);
      if (idx === -1) return { files: [...state.files, file] };
      const next = [...state.files];
      next[idx] = file;
      return { files: next };
    }),
  reset: () => set({ files: [], message: null }),
  message: null,
  setMessage: (message) => set({ message }),
}));
