import { RouteModule, HTTPContext } from "xerus";
import { Layout } from "../../src/layouts";
import { Header, MenuMain, Overlay } from "../../src/components";


let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  return c.jsx(
    <Layout title='Phillip England - Software Developer'>
        <>
            <Header mainText="About" subText="Phillip England" hasIcon={true}/>
            <MenuMain />
            <Overlay/>
            <div className="border-b">
                <img src='/static/gray-me.png'></img>

            </div>
        </>
    </Layout>
  );
});

export default module;