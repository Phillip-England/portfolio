import { InitModule } from "xerus";
import { logger, Xerus } from "xerus";

let module = new InitModule();

module.init(async (app: Xerus) => {
  app.use(logger);
  app.static("static");
});

export default module;
