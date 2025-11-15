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
        <div className="grid grid-cols-1 gap-8 md:grid-cols-2 p-4">
          <ProjectCard name='rlex' href="/projects/rlex" languages={["rust"]} description="A cursor-based lexer for parsing and tokenizing utf-8 string." />
          <ProjectCard name='xerus' href="/projects/xerus" languages={["bun", "ts"]} description="An express-inspired web framework with it's own file-based routing system." />
          <ProjectCard name='gtml' href="/projects/gtml" languages={["go"]} description="Make writing html in Go a breeze" />
          <ProjectCard name='finli' href="/projects/finli" languages={["rust"]} description="A system for generating .pdf invoices from a directory of .pdf receipts" />
          <ProjectCard name='translation-bot' href="/projects/translation-bot" languages={["go"]} description="Setup an API endpoint for auto-translating messages on GroupMe" />
          <ProjectCard name='bible-bot' href="/projects/bible-bot" languages={["go"]} description="Webscrape multiple translations of the bible from https://bible.com and save each verse in a Sqlite database" />
          <ProjectCard name='wherr' href="/projects/wherr" languages={["go"]} description="Never wonder 'wherr' your error occured again" />
          {/* <ProjectCard href="/projects/xerus" languages={["rust"]} /> */}
        </div>
        <script src="/static/three.js"></script>
      </>
    </Layout>,
  );
});

export default module;
