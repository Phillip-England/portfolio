

export const TitleLinks = () => {
  let titleLinkHtml = `<title-links target="#tw-markdown" link-class="text-sm text-gray-600 hover:text-blue-500 my-toc-link" link-wrapper-class="py-1 border-l border-gray-200 pl-2 my-toc-wrapper" offset="-190"></title-links>`
  return (
    <div className="w-full" dangerouslySetInnerHTML={{ __html: titleLinkHtml }}/>
  )
}