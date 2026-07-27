# Optional S3 storage

Standalone module implementing `storage.Storage` with S3 multipart uploads.
It is **not** part of the default backend module — delete this folder if you do not need S3.

## Use in your app

```go
import s3storage "github.com/file-uploads-go/backend/optional/s3storage"

uploader, err := s3storage.New(os.Getenv("S3_BUCKET"))
// pass uploader as upload.Options.Storage
```

With a local replace in your `go.mod`:

```
require github.com/file-uploads-go/backend/optional/s3storage v0.0.0
replace github.com/file-uploads-go/backend/optional/s3storage => ./optional/s3storage
```

Then:

```bash
cd optional/s3storage && go mod tidy
```

Set `S3_BUCKET` and ensure AWS credentials are available via the default credential chain.
