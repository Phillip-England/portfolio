

export const Header = (props: {
    mainText: string,
    subText: string,
    hasIcon: boolean,
}) => {
    return (
        <header className="p-4 flex flex-row gap-4 justify-between items-center top-0 sticky bg-white z-50 select-none">
            <div className="flex flex-col gap-2">
                <h1 className="font-mono text-3xl">{props.mainText}</h1>
                <p className="font-mono">{props.subText}</p>
            </div>
            { props.hasIcon ?
                <>
                    <svg id='icon-bars' className="w-8 h-8 flex" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24">
                        <path stroke="currentColor" strokeLinecap="round" strokeWidth="2" d="M5 7h14M5 12h14M5 17h14"/>
                    </svg>
                    <svg id='icon-x' className="w-8 h-8 hidden" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24">
                        <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18 17.94 6M18 18 6.06 6"/>
                    </svg>
                </>

                :
                <></>
            }
        </header>
    )
}


export const MenuMain = (props: {

}) => {
    return (
        <nav id='menu-main' className="bg-white p-4 text-lg w-[60%] max-w-sm h-full fixed left-0 top-0 z-30 hidden">
            <ul className="flex flex-col gap-2">
                <div className="h-[90px]"></div>
                <NavLink href="/" text="Home" />
                <NavLink href="/about" text="About" />
                <NavLink href="/projects" text="Projects" />
                <NavLink href="/blog" text="Blog" />
            </ul>
        </nav>
    )
}


export const NavLink = (props: {
    text: string,
    href: string
}) => {
    return (
        <li className="p-2">
            <a href={props.href}>{props.text}</a>
        </li>
    )
}

export const Overlay = (props: {

}) => {
    return (
        <div id='overlay' className="fixed top-0 h-screen w-screen bg-black opacity-40 hidden"></div>
    )
}