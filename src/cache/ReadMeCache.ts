

class ReadMeCache {
  visitedUrls: Record<string, string> = {}
  constructor() {}
  save(url: string, readMeHtml: string) {
    console.log('cache saved')
    this.visitedUrls[url] = readMeHtml
  }
  load(url: string): string | null {
    let potentialUrl = this.visitedUrls[url]
    if (potentialUrl) {
      console.log('cache loaded')
      return potentialUrl
    }
    return null
  }
}

export let readMeCache = new ReadMeCache()