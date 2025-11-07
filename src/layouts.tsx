import type { JSX } from "react"



export const Layout = (props: {
    title: string,
    children: JSX.Element
}) => {
    return (
        <html lang="en">
        <head>
            <meta charSet="UTF-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1.0" />
            <link href='/static/output.css' rel='stylesheet'></link>
            <title>{props.title}</title>
        </head>
        <body className="font-mono">
            {props.children}
            <script src="/static/index.js"></script>
        </body>
        </html>
    )
}