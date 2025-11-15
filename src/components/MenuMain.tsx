import { NavLink } from "./NavLink";

export const MenuMain = (props: {
  reqPath: string;
}) => {
  return (
    <nav
      id="menu-main"
      className="bg-white p-4 text-lg w-[70%] max-w-sm h-full fixed left-0 top-0 z-30 hidden"
    >
      <ul className="flex flex-col gap-2">
        <div className="h-[90px]"></div>
        <NavLink href="/" text="Home" reqPath={props.reqPath} />
        <NavLink href="/about" text="About" reqPath={props.reqPath} />
        <NavLink href="/projects" text="Projects" reqPath={props.reqPath} />
        <NavLink href="/blog" text="Blog" reqPath={props.reqPath} />
      </ul>
    </nav>
  );
};
