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
        <body className="font-mono flex justify-center">
            <div id='root' className="w-full md:w-[80%] max-w-[900px]">
                {props.children}
                <footer className="p-12 flex items-center justify-center text-gray-500">
                  <p className="text-center">This site was generate using <a className="underline text-primary" href='https://github.com/phillip-england/flint' target="_blank">flint</a>, a language-agnostic static-site generator I developed.</p>
                </footer>
            </div>
            <script src="/static/index.js"></script>
        </body>
        </html>
    )
}