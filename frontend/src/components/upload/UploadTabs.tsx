import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ChunkedUploadPanel } from "@/components/upload/ChunkedUploadPanel";
import { MultiChunkedUploadPanel } from "@/components/upload/MultiChunkedUploadPanel";
import { StreamUploadPanel } from "@/components/upload/StreamUploadPanel";
import { useUploadStore } from "@/hooks/useUploadStore";

export function UploadTabs() {
  const { activeTab, setActiveTab, message, reset } = useUploadStore();

  return (
    <Card>
      <CardHeader>
        <CardTitle>File upload modes</CardTitle>
        <CardDescription>
          All modes write to local disk on the Go backend. Detached clients live in{" "}
          <code className="rounded bg-zinc-100 px-1">upload-lib/</code>.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs
          value={activeTab}
          onValueChange={(tab) => {
            reset();
            setActiveTab(tab);
          }}
        >
          <TabsList>
            <TabsTrigger value="stream">Streamed</TabsTrigger>
            <TabsTrigger value="stream-multi">Streamed multi</TabsTrigger>
            <TabsTrigger value="chunked">Chunked</TabsTrigger>
            <TabsTrigger value="chunked-multi">Chunked multi</TabsTrigger>
          </TabsList>
          <TabsContent value="stream">
            <p className="mb-3 text-sm text-zinc-600">
              Single-file multipart stream with XHR progress.
            </p>
            <StreamUploadPanel />
          </TabsContent>
          <TabsContent value="stream-multi">
            <p className="mb-3 text-sm text-zinc-600">
              Multiple files in one multipart POST; aggregate byte progress.
            </p>
            <StreamUploadPanel multiple />
          </TabsContent>
          <TabsContent value="chunked">
            <p className="mb-3 text-sm text-zinc-600">
              Resumable 5MB chunks with concurrency and retries.
            </p>
            <ChunkedUploadPanel />
          </TabsContent>
          <TabsContent value="chunked-multi">
            <p className="mb-3 text-sm text-zinc-600">
              Parallel chunked sessions per file (2 files × 3 chunks).
            </p>
            <MultiChunkedUploadPanel />
          </TabsContent>
        </Tabs>
        {message ? (
          <p className="mt-4 rounded-md bg-zinc-50 px-3 py-2 text-sm text-zinc-700">{message}</p>
        ) : null}
      </CardContent>
    </Card>
  );
}
