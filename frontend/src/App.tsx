import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { UploadTabs } from "@/components/upload/UploadTabs";

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-screen bg-gradient-to-b from-zinc-100 to-zinc-200">
        <main className="mx-auto max-w-3xl px-4 py-10">
          <header className="mb-8">
            <p className="text-sm font-medium uppercase tracking-wide text-emerald-700">
              file-uploads-go
            </p>
            <h1 className="mt-1 text-3xl font-semibold tracking-tight text-zinc-900">
              Upload POC
            </h1>
            <p className="mt-2 text-zinc-600">
              React 19 frontend against a chi-backed Go upload service.
            </p>
          </header>
          <UploadTabs />
        </main>
      </div>
    </QueryClientProvider>
  );
}
