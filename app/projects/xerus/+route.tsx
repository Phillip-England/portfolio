import { HTTPContext, RouteModule } from "xerus";
import { loadProjectReadme } from "../../../src/cache/ReadMeCache";
import { ProjectPage } from "../../../src/components/ProjectPage";

let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  let projectName = "Xerus";
  let readMeHtml = await loadProjectReadme(projectName);
  return c.jsx(
    <ProjectPage
      projectName={projectName}
      reqPath={c.path}
      readMeHtml={readMeHtml}
    />,
  );
});

export default module;
