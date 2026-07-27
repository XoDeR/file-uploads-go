import { describe, expect, it } from "vitest";
import { expectedChunkSize, totalChunks } from "./chunk-math";

describe("totalChunks", () => {
  it("exact multiple", () => {
    expect(totalChunks(15, 5)).toBe(3);
  });
  it("remainder last chunk", () => {
    expect(totalChunks(12, 5)).toBe(3);
  });
  it("single chunk exact", () => {
    expect(totalChunks(5, 5)).toBe(1);
  });
  it("single chunk smaller", () => {
    expect(totalChunks(3, 5)).toBe(1);
  });
  it("large file", () => {
    expect(totalChunks(100 * 1024 * 1024 + 1, 5 * 1024 * 1024)).toBe(21);
  });
  it("rejects zero total", () => {
    expect(() => totalChunks(0, 5)).toThrow(/invalid size/);
  });
  it("rejects negative total", () => {
    expect(() => totalChunks(-1, 5)).toThrow(/invalid size/);
  });
  it("rejects zero chunk", () => {
    expect(() => totalChunks(10, 0)).toThrow(/invalid size/);
  });
  it("rejects negative chunk", () => {
    expect(() => totalChunks(10, -5)).toThrow(/invalid size/);
  });
});

describe("expectedChunkSize", () => {
  it("middle chunk", () => {
    expect(expectedChunkSize(12, 5, 0, 3)).toBe(5);
  });
  it("second chunk", () => {
    expect(expectedChunkSize(12, 5, 1, 3)).toBe(5);
  });
  it("last remainder", () => {
    expect(expectedChunkSize(12, 5, 2, 3)).toBe(2);
  });
  it("exact last", () => {
    expect(expectedChunkSize(15, 5, 2, 3)).toBe(5);
  });
  it("single chunk", () => {
    expect(expectedChunkSize(3, 5, 0, 1)).toBe(3);
  });
});
