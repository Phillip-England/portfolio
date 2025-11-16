import { FileRouter } from "xerus";
import path from "path";
import { primeReadMeCache } from "./src/cache/ReadMeCache";
import { ArticleCache } from "./src/cache/ArticleCache";


// await primeReadMeCache()

export const articleCache = await  ArticleCache.new(path.join(process.cwd(), "articles"))

let router = await FileRouter.new({
  "src": path.join(process.cwd(), "app"),
  "port": 8080,
});
await router.listen();
