import { HTTPContext, RouteModule } from "xerus";
import { Header } from "../../src/components/Header";
import { MenuMain } from "../../src/components/MenuMain";
import { Overlay } from "../../src/components/Overlay";
import { Layout } from "../../src/layouts/Layout";
import { TitleCard } from "../../src/components/TitleCard";
import { ProjectCard } from "../../src/components/ProjectCard";

let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  return c.jsx(
    <Layout title="Projects - Phillip England">
      <>
        <Header
          mainText="Projects"
          subText="Phillip England"
          hasIcon={true}
          hasNav={true}
          reqPath={c.path}
        />
        <MenuMain reqPath={c.path} />
        <Overlay />
        <div className="flex items-center justify-center p-12">
          <div id="three-screen" className="h-[200px] w-full"></div>
        </div>
        <div className="p-4 md:max-w-lg">
          <TitleCard title="I'm Always Building Something">
            <p>
              Over the years, I've developed multiple tools which help me in
              both my personal and professional life. I'm always looking for
              opportunities to solve problems with software.
            </p>
          </TitleCard>
        </div>
        <div className="flex flex-col p-4">
          <ProjectCard href="/projects/rlex" />
          <ProjectCard href="/projects/xerus" />
        </div>
        <script src="/static/three.js"></script>
      </>
    </Layout>,
  );
});

export default module;
