import { BodyType, HTTPContext, RouteModule } from "xerus";
import { Header } from "../../src/components/Header";
import { MenuMain } from "../../src/components/MenuMain";
import { Overlay } from "../../src/components/Overlay";
import { Layout } from "../../src/layouts/Layout";
import { TitleCard } from "../../src/components/TitleCard";
import { ProjectCard } from "../../src/components/ProjectCard";
import { RandomBeads } from "../../src/components/RandomBeads";
import { contactFormTable } from "../..";


let module = new RouteModule();

module.get(async (c: HTTPContext) => {
  return c.jsx(
    <Layout title="Projects - Phillip England">
      <>
        <Header
          mainText="Contact"
          subText="Phillip England"
          hasIcon={true}
          hasNav={true}
          reqPath={c.path}
        />
        <MenuMain reqPath={c.path} />
        <Overlay />
        <div className="flex items-center flex-col gap-12 p-2 lg:p-12">
          <form
            action="/contact"
            method="POST"
            className="max-w-[700px] flex flex-col gap-8 p-8 w-full md:max-w-[400px]"
          >
            <h2 className="text-2xl">Contact Me</h2>
            <div className="flex flex-col gap-6">
              <input
                name="name"
                data-slot="input"
                className="file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground border-input h-9 w-full min-w-0 rounded-md border bg-transparent border-gray-200 px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[1px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive"
                placeholder="Name"
                type="text"
              >
              </input>
              <input
                name="email"
                data-slot="input"
                className="file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground border-input h-9 w-full min-w-0 rounded-md border bg-transparent border-gray-200 px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[1px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive"
                placeholder="Email"
                type="email"
              >
              </input>
              <textarea
                name="message"
                data-slot="textarea"
                className="border-input border-gray-200 placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:bg-input/30 flex field-sizing-content min-h-16 w-full rounded-md border bg-transparent px-3 py-2 text-base shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-[1px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
                placeholder="Type your message here."
              >
              </textarea>
              <button
                data-slot="button"
                className="inline-flex bg-black text-white cursor-pointer items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 [&amp;_svg]:pointer-events-none [&amp;_svg:not([class*='size-'])]:size-4 shrink-0 [&amp;_svg]:shrink-0 outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:bg-input/30 dark:border-input dark:hover:bg-input/50 h-9 px-4 py-2 has-[&gt;svg]:px-3"
                type="submit"
              >
                Submit
              </button>
            </div>
          </form>
          <RandomBeads />
        </div>
      </>
    </Layout>,
  );
});

module.post(async (c: HTTPContext) => {
  let form = await c.parseBody(BodyType.FORM);
  let name = form["name"];
  if (!name) {
  }
  let email = form["email"];
  if (!email) {
  }
  let message = form["message"];
  if (!message) {
  }
  contactFormTable.insert(name, email, message)
  return c.redirect("/contact", 302);
});

export default module;
