import { RouteModule, HTTPContext } from "xerus";
import { Layout } from "../../src/layouts";
import { Header, MenuMain, Overlay } from "../../src/components";


let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  return c.jsx(
    <Layout title='Phillip England - Software Developer'>
        <>
            <Header mainText="About" subText="Phillip England" hasIcon={true} hasNav={true} reqPath={c.path}/>
            <MenuMain reqPath={c.path} />
            <Overlay/>
            <div className="flex items-center justify-center">
                <img src='/static/profile.png' className=""></img>
            </div>
            <article className="md:max-w-lg p-4 gap-12 flex flex-col">
              <div className="flex flex-col gap-8">
                <h2 className="text-2xl">10 Years at Chick-fil-A</h2>
                <p>I currently serve as the Senior Director of Financial Stewardship with Chick-fil-A Southroads and Chick-fil-A Utica out of Tulsa Okalhoma. I partner directly with David Chen at a combined volume of over 14 million dollars in sales per year. I am directly responsible for the financial success of the organization and partner with a team of Directors to manage costs and ensure consistent profits.</p>
              </div>
              <div className="flex flex-col gap-8">
                <h2 className="text-2xl">Self-Taught Developer</h2>
                <p>Seven years ago, I thought to myself, "I want to build a website." Seven years later, I am still tumbling down the rabbit hole. Software is my life skill. It is the thing I will continually polish throughout the course of my life due to passion, irrespective of my career. I do it for the love of the craft.</p>
              </div>
              <div className="flex flex-col gap-8">
                <h2 className="text-2xl">Always Improving</h2>
                <p>I always have a passion or skill I am actively investing my freetime into. Some people call it obsession; I call it accountability. I expect the best out of myself in everything I do. I expect the future version of myself to be more knowledable than I am today. These motivation are derived from within and have consistently held me accountable throughout my life. It is the reason I apologize when I am wrong and why I am always honing my craft. It is what I <i>ought</i> to do.</p>
              </div>
            </article>
        </>
    </Layout>
  );
});

export default module;