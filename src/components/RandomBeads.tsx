export const RandomBeads = () => {
  let component = '<random-beads count="16"></tw-markdown>';
  return <div dangerouslySetInnerHTML={{ __html: component }} />;
};
