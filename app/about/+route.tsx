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
            <Header mainText="About" subText="Phillip England" hasIcon={true} hasNav={true} reqPath={c.path}/>
            <MenuMain reqPath={c.path} />
            <Overlay/>
            <div className="flex items-center justify-center">
                <img src='/static/profile.png' className=""></img>
            </div>
            <article className="md:max-w-lg p-4 gap-12 flex flex-col">
              <TitleCard title='10 Years at Chick-fil-A'>
                <p>I currently serve as the Senior Director of Financial Stewardship with <a className="text-primary underline" href='https://www.chick-fil-a.com/locations/ok/southroads-shopping-center' target='_blank'>Chick-fil-A Southroads</a> and <a className='text-primary underline' href='https://www.chick-fil-a.com/locations/ok/13th-utica' target='_blank'>Chick-fil-A Utica</a> out of Tulsa, Okalhoma. I partner directly with <a className='text-primary underline' href='https://www.chenchiasin.com/' target='_blank'>David Chen</a> at a combined volume of over 14 million dollars in sales per year. I am directly responsible for the financial success of the organization and partner with a team of Directors to manage costs and ensure consistent profits.</p>
              </TitleCard>
              <TitleCard title='Self-Taught Developer'>
                <p>Seven years ago, I thought to myself, "I want to build a website." Seven years later, I am still tumbling down the rabbit hole. Software is my life skill. It is the thing I will continually polish throughout the course of my life due to passion, irrespective of my career. I do it for the love of the craft.</p>
              </TitleCard>
              <TitleCard title='A Love of Tooling'>
                <p>I started my developer journey in client-side development. My first real project was an RBG controller which let you change the background color of the webpage. This is where I honed in my HTML, Javascript, and CSS skills. Eventually, I found myself on the server. I know Typescript (Node/Bun), Python, Go, Rust, and even dabbled in bit of Zig/C. I am most proficeint in Golang and Typescript. In my personal time, I like to develop tools revolving around the web devloper ecosystem. My dream project is some sort of compiler-driven UI framework/DSL. I draw major inspiration from projects like Tailwind, HTMX, Svelte, and even Hugo.</p>
              </TitleCard>
            </article>
        </>
    </Layout>
  );
});

export default module;