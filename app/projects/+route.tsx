import { RouteModule, HTTPContext } from "xerus";
import { Header } from "../../src/components/Header";
import { MenuMain } from "../../src/components/MenuMain";
import { Overlay } from "../../src/components/Overlay";
import { Layout } from "../../src/layouts/Layout";
import { TitleCard } from "../../src/components/TitleCard";

let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  return c.jsx(
    <Layout title='Phillip England - Software Developer'>
        <>
            <Header mainText="Projects" subText="Phillip England" hasIcon={true} hasNav={true} reqPath={c.path}/>
            <MenuMain reqPath={c.path} />
            <Overlay/>
            <div className="flex items-center justify-center p-12">
              <div id='three-screen' className="h-[200px] w-full"></div>
            </div>
            <div className="p-4 md:max-w-lg">
              <TitleCard title="I'm Always Building Something">
                <p>Over the years, I've developed multiple tools which help me in both my personal and professional life. I'm always looking for opportunities to solve problems with software.</p>
              </TitleCard>
            </div>
            <div className="flex flex-col p-4">
              <div className="rounded-lg border border-gray-100  shadow-md w-fit p-4 flex flex-col gap-4 hover:shadow-xl hover:shadow-secondary cursor-pointer">
                <div className="flex flex-row justify-between items-center">
                  <h2 className="text-xl font-semibold text-gray-500">rlex</h2>
                  <img src='/static/ferris.svg' className="w-10"></img>
                </div>
                <p className="text-sm max-w-sm">A cursor-based lexer for parsing and tokenizing utf-8 string.</p>
              </div>
            </div>
            <script src='/static/three.js'></script>
        </>
    </Layout>
  );
});

export default module;