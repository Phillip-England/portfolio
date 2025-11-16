import { readdir } from "fs/promises";
import path from "path";
import { Article } from "./Article";
import { JSX } from "react";

export class ArticleCache {
  articleDirPath: string;
  cache: Record<string, Article>;
  constructor(articleDirPath: string, cache: Record<string, Article>) {
    this.articleDirPath = articleDirPath;
    this.cache = cache;
  }
  static async new(articleDirPath: string) {
    let files = await readdir(articleDirPath, {
      recursive: true,
    });
    let cache: Record<string, Article> = {};
    for (const file of files) {
      let article = await Article.new(path.join(articleDirPath, file));
      cache[article.number] = article;
    }
    return new ArticleCache(articleDirPath, cache);
  }
  load(number: string) {
    let article = this.cache[number];
    if (article) {
      return article;
    }
    throw new Error(`no aricle with number ${number}`);
  }
}
