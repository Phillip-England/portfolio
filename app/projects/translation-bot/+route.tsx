import { HTTPContext, RouteModule } from "xerus";
import { ProjectPage } from "../../../src/components/ProjectPage";
import { loadProjectReadme } from "../../../src/cache/ReadMeCache";
loadProjectReadme;

let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  let projectName = "Translation-Bot";
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
