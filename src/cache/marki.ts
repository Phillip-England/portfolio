import { $ } from "bun";

export async function markdownToHtml(path: string) {
  let result = await $`marki convert dracula < ${path}`.quiet();
  let html = result.stdout.toString().trim()
  return html
}