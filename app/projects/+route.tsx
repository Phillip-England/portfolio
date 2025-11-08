import { RouteModule, HTTPContext } from "xerus";
import { Header } from "../../src/components/Header";
import { MenuMain } from "../../src/components/MenuMain";
import { Overlay } from "../../src/components/Overlay";
import { Layout } from "../../src/layouts/Layout";

let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  return c.jsx(
    <Layout title='Phillip England - Software Developer'>
        <>
            <Header mainText="About" subText="Phillip England" hasIcon={true} hasNav={true} reqPath={c.path}/>
            <MenuMain reqPath={c.path} />
            <Overlay/>
            <div className="flex items-center justify-center p-12">
              <div id='three-screen' className="h-[200px] w-full"></div>
            </div>
            <script src='/static/three.js'></script>
        </>
    </Layout>
  );
});

export default module;