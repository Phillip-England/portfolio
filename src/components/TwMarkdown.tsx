export const TwMarkdown = (props: {
  html: string;
}) => {
  let component = "<tw-markdown>" + props.html + "</tw-markdown>";
  return (
    <div id="tw-markdown" dangerouslySetInnerHTML={{ __html: component }} />
  );
};
