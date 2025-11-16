import { TwMarkdown } from "./TwMarkdown";
import { Layout } from "../layouts/Layout";
import { Header } from "./Header";
import { MenuMain } from "./MenuMain";
import { Overlay } from "./Overlay";
import { TitleLinks } from "./TitleLinks";
import { loadProjectReadme } from "../cache/ReadMeCache";

export const ProjectPage = (props: {
  reqPath: string;
  projectName: string;
  readMeHtml: string;
}) => {
  return (
    <Layout title={props.projectName + " - Phillip England"}>
      <>
        <Header
          mainText={props.projectName}
          subText="Phillip England"
          hasIcon={true}
          hasNav={true}
          reqPath={props.reqPath}
        />
        <MenuMain reqPath={props.reqPath} />
        <Overlay />
        <div className="p-4 grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="md:col-span-2 col-span-3">
            <TwMarkdown html={props.readMeHtml} />
          </div>
          <div className="md:col-span-1 hidden md:flex relative">
            <div className="fixed top-50 z-30 bg-white overflow-y-scroll h-[500px] min-w-[270px]">
              <TitleLinks />
            </div>
          </div>
        </div>
      </>
    </Layout>
  );
};
