import { FileRouter } from "xerus";
import path from "path";
import { primeReadMeCache } from "./src/cache/ReadMeCache";


await primeReadMeCache()

let router = await FileRouter.new({
  "src": path.join(process.cwd(), "app"),
  "port": 8080,
});
await router.listen();
