import { FileRouter } from "xerus";
import path from "path";
import { primeReadMeCache } from "./src/cache/ReadMeCache";
import { ArticleCache } from "./src/cache/ArticleCache";
import { Database } from "bun:sqlite";
import { ContactFormTable } from "./src/database/SqLiteDb";

await primeReadMeCache()
export const articleCache = await ArticleCache.new(
  path.join(process.cwd(), "articles"),
);

export const db = new Database(path.join(process.cwd(), "main.db"));
export const contactFormTable =  new ContactFormTable(db)

let router = await FileRouter.new({
  "src": path.join(process.cwd(), "app"),
  "port": 8080,
});
await router.listen();
