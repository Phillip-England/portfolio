export const NavLink = (props: {
    text: string,
    href: string,
    reqPath: string,
}) => {
    let activeClass = 'p-2 md:p-0 underline'
    if (props.reqPath == props.href) {
      activeClass = activeClass +  ' text-[#f802fa]'
    }
    return (
        <li className={activeClass}>
            <a href={props.href}>{props.text}</a>
        </li>
    )
}