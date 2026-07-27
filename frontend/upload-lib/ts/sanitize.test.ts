import { describe, expect, it } from "vitest";
import { sanitizeFilename } from "./sanitize";

describe("sanitizeFilename", () => {
  it("normal", () => {
    expect(sanitizeFilename("photo.jpg")).toBe("photo.jpg");
  });
  it("path traversal unix", () => {
    expect(sanitizeFilename("../../etc/passwd")).toBe("passwd");
  });
  it("path traversal nested", () => {
    expect(sanitizeFilename("foo/../../bar.txt")).toBe("bar.txt");
  });
  it("dangerous chars", () => {
    expect(sanitizeFilename('a:b*c?d|e<f>g"h.txt')).toBe("abcdefgh.txt");
  });
  it("null byte", () => {
    expect(sanitizeFilename("file\x00name.txt")).toBe("filename.txt");
  });
  it("dots only after sanitize", () => {
    expect(sanitizeFilename("..")).toBe("unnamed_file");
  });
  it("empty", () => {
    expect(sanitizeFilename("")).toBe("unnamed_file");
  });
  it("dot", () => {
    expect(sanitizeFilename(".")).toBe("unnamed_file");
  });
  it("backslash path", () => {
    expect(sanitizeFilename("..\\..\\secret.txt")).toBe("secret.txt");
  });
});
