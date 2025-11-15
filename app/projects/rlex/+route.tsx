import { HTTPContext, RouteModule } from "xerus";
import { Header } from "../../../src/components/Header";
import { MenuMain } from "../../../src/components/MenuMain";
import { Overlay } from "../../../src/components/Overlay";
import { Layout } from "../../../src/layouts/Layout";
import { loadProjectReadme } from "../../../src/cache/ReadMeCache";


let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  let readMeHtml = await loadProjectReadme("rlex");
  readMeHtml = '<tw-markdown>'+readMeHtml+"</tw-markdown>"
  let titleLinkHtml = `<title-links target="#readme-content" link-class="text-sm text-gray-600 hover:text-blue-500 my-toc-link" link-wrapper-class="py-1 border-l border-gray-200 pl-2 my-toc-wrapper" offset="-190"></title-links>`
  // console.log(readMeCache)
  return c.jsx(
    <Layout title="rlex - Phillip England">
      <>
        <Header
          mainText="rlex"
          subText="Phillip England"
          hasIcon={true}
          hasNav={true}
          reqPath={c.path}
        />
        <MenuMain reqPath={c.path} />
        <Overlay />
        {/* <div className="flex items-center justify-center p-12">
          <div id="three-screen" className="h-[200px] w-full"></div>
        </div> */}
        <div className="p-4 grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="md:col-span-2 col-span-3">
            <div id="readme-content" dangerouslySetInnerHTML={{ __html: readMeHtml }}/>
          </div>
          <div className="md:col-span-1 hidden md:flex relative">
            <div className="fixed top-50">
              <div dangerouslySetInnerHTML={{ __html: titleLinkHtml }}/>
            </div>
          </div>
        </div>
        <script src="/static/three.js"></script>
      </>
    </Layout>,
  );
});

export default module;