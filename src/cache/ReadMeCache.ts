import { $ } from "bun";

class ProjectTimestamp {
  readMeHtml: string;
  lastFetched: Date;
  constructor(readMeHtml: string) {
    this.readMeHtml = readMeHtml;
    this.lastFetched = new Date();
  }
}

class ReadMeCache {
  visitedUrls: Record<string, ProjectTimestamp> = {};
  save(url: string, readMeHtml: string) {
    this.visitedUrls[url] = new ProjectTimestamp(readMeHtml);
  }
  load(url: string): ProjectTimestamp | null {
    let potentialCacheEntry = this.visitedUrls[url];
    if (potentialCacheEntry) {
      return potentialCacheEntry;
    }
    return null;
  }
}

export let readMeCache = new ReadMeCache();
const CACHE_EXPIRATION_MS = 5 * 60 * 1000;

export async function loadProjectReadme(projectName: string): Promise<string> {
  let readMeUrl =
    `https://raw.githubusercontent.com/Phillip-England/${projectName}/refs/heads/main/README.md`;
  let potentialCacheEntry = readMeCache.load(readMeUrl);
  if (potentialCacheEntry) {
    const now = new Date().getTime();
    const lastFetched = potentialCacheEntry.lastFetched.getTime();
    if (now - lastFetched < CACHE_EXPIRATION_MS) {
      return potentialCacheEntry.readMeHtml;
    }
  }
  let res = await fetch(readMeUrl);
  if (res.status == 200) {
    let readMeText = await res.text();
    const readMeResponse = new Response(readMeText);
    let result = await $`marki convert dracula < ${readMeResponse}`.quiet();
    let html = result.stdout.toString().trim();
    readMeCache.save(readMeUrl, html);
    return html;
  }
  return `<p>unable to load README.md from ${readMeUrl}</p>`;
}


