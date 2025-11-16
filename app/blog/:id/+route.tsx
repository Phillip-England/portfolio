import { HTTPContext, RouteModule } from "xerus";
import { Header } from "../../../src/components/Header";
import { MenuMain } from "../../../src/components/MenuMain";
import { Overlay } from "../../../src/components/Overlay";
import { Layout } from "../../../src/layouts/Layout";
import { TitleCard } from "../../../src/components/TitleCard";
import { articleCache } from "../../..";
import { Article } from "../../../src/cache/Article";
import { TwMarkdown } from "../../../src/components/TwMarkdown";
import { TitleLinks } from "../../../src/components/TitleLinks";

let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  let path = c.path;
  let pathParts = path.split("/");
  let articleNumber = pathParts[pathParts.length - 1];
  let article: Article;
  try {
    article = articleCache.load(articleNumber);
  } catch (e: any) {
    console.error(e.message);
    c.setStatus(404);
    return c.html("<h1>404 not found</h1>");
  }
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
        <div className="p-4 flex flex-col gap-4">
          <h2 className="text-3xl">{article.name}</h2>
          <div className="flex flex-col gap-1">
            <p>{article.subText}</p>
            <p className="text-sm text-gray-500">{article.dob}</p>
          </div>
        </div>
        <div className="p-4 grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="md:col-span-2 col-span-3">
            <TwMarkdown html={article.html} />
          </div>
          <div className="md:col-span-1 hidden md:flex relative">
            <div className="fixed top-50 z-30 bg-white overflow-y-scroll h-[500px]">
              <TitleLinks />
            </div>
          </div>
        </div>
      </>
    </Layout>,
  );
});

export default module;
