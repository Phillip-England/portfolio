import { NavLink } from "./NavLink";

export const Header = (props: {
  mainText: string;
  subText: string;
  hasIcon: boolean;
  hasNav: boolean;
  reqPath: string;
}) => {
  let theBlinkerHtml = '<the-blinker>_</the-blinker>'
  return (
    <header className="p-4 flex gap-8 justify-between top-0 sticky bg-white z-50 select-none flex-row md:flex-col">
      <div className="flex flex-col gap-2">
        <h1 className="font-mono text-3xl">{props.mainText}</h1>
        <p className="font-mono">{props.subText}
          <span dangerouslySetInnerHTML={{ __html: theBlinkerHtml }}/>
        </p>
      </div>
      {props.hasIcon
        ? (
          <div className="flex items-center md:hidden">
            <svg
              id="icon-bars"
              className="w-8 h-8 flex md:hidden"
              aria-hidden="true"
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              fill="none"
              viewBox="0 0 24 24"
            >
              <path
                stroke="currentColor"
                strokeLinecap="round"
                strokeWidth="2"
                d="M5 7h14M5 12h14M5 17h14"
              />
            </svg>
            <svg
              id="icon-x"
              className="w-8 h-8 hidden md:hidden"
              aria-hidden="true"
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              fill="none"
              viewBox="0 0 24 24"
            >
              <path
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M6 18 17.94 6M18 18 6.06 6"
              />
            </svg>
          </div>
        )
        : <></>}
      {props.hasNav
        ? (
          <nav className="hidden md:flex flex-row justify-between items-center">
            <ul className="flex flex-row gap-8">
              <NavLink href="/" text="Home" reqPath={props.reqPath} />
              <NavLink href="/about" text="About" reqPath={props.reqPath} />
              <NavLink href="/contact" text="Contact" reqPath={props.reqPath} />
              <NavLink
                href="/projects"
                text="Projects"
                reqPath={props.reqPath}
              />
              <NavLink href="/blog" text="Blog" reqPath={props.reqPath} />
            </ul>
            <ul className="flex flex-row gap-8">
              <li>
                <a
                  href="https://github.com/phillip-england"
                  className="cursor-pointer"
                >
                  <svg
                    className="w-12 h-12 text-gray-800"
                    aria-hidden="true"
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      fillRule="evenodd"
                      d="M12.006 2a9.847 9.847 0 0 0-6.484 2.44 10.32 10.32 0 0 0-3.393 6.17 10.48 10.48 0 0 0 1.317 6.955 10.045 10.045 0 0 0 5.4 4.418c.504.095.683-.223.683-.494 0-.245-.01-1.052-.014-1.908-2.78.62-3.366-1.21-3.366-1.21a2.711 2.711 0 0 0-1.11-1.5c-.907-.637.07-.621.07-.621.317.044.62.163.885.346.266.183.487.426.647.71.135.253.318.476.538.655a2.079 2.079 0 0 0 2.37.196c.045-.52.27-1.006.635-1.37-2.219-.259-4.554-1.138-4.554-5.07a4.022 4.022 0 0 1 1.031-2.75 3.77 3.77 0 0 1 .096-2.713s.839-.275 2.749 1.05a9.26 9.26 0 0 1 5.004 0c1.906-1.325 2.74-1.05 2.74-1.05.37.858.406 1.828.101 2.713a4.017 4.017 0 0 1 1.029 2.75c0 3.939-2.339 4.805-4.564 5.058a2.471 2.471 0 0 1 .679 1.897c0 1.372-.012 2.477-.012 2.814 0 .272.18.592.687.492a10.05 10.05 0 0 0 5.388-4.421 10.473 10.473 0 0 0 1.313-6.948 10.32 10.32 0 0 0-3.39-6.165A9.847 9.847 0 0 0 12.007 2Z"
                      clipRule="evenodd"
                    />
                  </svg>
                </a>
              </li>
              <li>
                <a href="https://x.com/Phillip98282955">
                  <svg
                    className="w-12 h-12 text-gray-800"
                    aria-hidden="true"
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M13.795 10.533 20.68 2h-3.073l-5.255 6.517L7.69 2H1l7.806 10.91L1.47 22h3.074l5.705-7.07L15.31 22H22l-8.205-11.467Zm-2.38 2.95L9.97 11.464 4.36 3.627h2.31l4.528 6.317 1.443 2.02 6.018 8.409h-2.31l-4.934-6.89Z" />
                  </svg>
                </a>
              </li>
              <li>
                <a href="https://www.facebook.com/phillip.england.559842/">
                  <svg
                    className="w-12 h-12 text-gray-800"
                    aria-hidden="true"
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      fillRule="evenodd"
                      d="M13.135 6H15V3h-1.865a4.147 4.147 0 0 0-4.142 4.142V9H7v3h2v9.938h3V12h2.021l.592-3H12V6.591A.6.6 0 0 1 12.592 6h.543Z"
                      clipRule="evenodd"
                    />
                  </svg>
                </a>
              </li>
            </ul>
          </nav>
        )
        : <></>}
    </header>
  );
};
