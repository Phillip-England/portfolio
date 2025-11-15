const IconRust = () => {
  return <img src="/static/ferris.svg" className="w-10"></img>;
};

const LanguageIcon = (langName: string) => {
  switch (langName) {
    case "rust":
      return <IconRust/>
  }    
};

export const ProjectCard = (props: {
  href: string;
}) => {
  return (
    <a
      href={props.href}
      className="rounded-lg border border-gray-100 shadow-md w-fit p-4 flex flex-col gap-4 hover:shadow-xl hover:shadow-secondary cursor-pointer"
    >
      <div className="flex flex-row justify-between items-center">
        <h2 className="text-xl font-semibold text-gray-500">rlex</h2>
        <img src="/static/ferris.svg" className="w-10"></img>
      </div>
      <p className="text-sm max-w-sm">
        A cursor-based lexer for parsing and tokenizing utf-8 string.
      </p>
    </a>
  );
};
