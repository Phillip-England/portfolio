import type { JSX } from "react"
import { Footer } from "../components/Footer"



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
                <Footer />
            </div>
            <script src="/static/index.js"></script>
        </body>
        </html>
    )
}