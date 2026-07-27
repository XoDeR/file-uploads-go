/** Mirrors Go validation.SanitizeFilename */
export function sanitizeFilename(filename: string): string {
  let name = filename.replace(/\\/g, "/");
  const slash = name.lastIndexOf("/");
  if (slash >= 0) {
    name = name.slice(slash + 1);
  }

  name = name
    .replace(/\.\./g, "")
    .replace(/\//g, "_")
    .replace(/\x00/g, "")
    .replace(/[<>:"|?*]/g, "");

  if (name === "" || name === ".") {
    return "unnamed_file";
  }
  return name;
}
