import { HTTPContext, RouteModule } from "xerus";
import { Header } from "../../src/components/Header";
import { MenuMain } from "../../src/components/MenuMain";
import { Overlay } from "../../src/components/Overlay";
import { Layout } from "../../src/layouts/Layout";
import { TitleCard } from "../../src/components/TitleCard";
import { ProjectCard } from "../../src/components/ProjectCard";
import { AllArticles } from "../../src/components/AllArticles";

let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  return c.jsx(
    <Layout title="Projects - Phillip England">
      <>
        <Header
          mainText="Blog"
          subText="Phillip England"
          hasIcon={true}
          hasNav={true}
          reqPath={c.path}
        />
        <MenuMain reqPath={c.path} />
        <Overlay />

        <div className="p-4 md:max-w-lg">
          <TitleCard title="See What I'm Learning">
            <p>
              As I progress in my skillset, I intend to continue posting
              articles about what I've learned.
            </p>
          </TitleCard>
        </div>
        <div className="p-4 flex flex-col gap-8">
          <h2 className="text-2xl">Articles</h2>
          <AllArticles />
        </div>
      </>
    </Layout>,
  );
});

export default module;
