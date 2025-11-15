
export const ProjectCard = (props: {
  name: string;
  href: string;
  languages: string[],
  description: string,
}) => {
  let langElement = ""
  for (let i = 0; i < props.languages.length; i++) {
    langElement += `<img class='w-8' src='/static/${props.languages[i]}.svg'></img>`
  }
  return (
    <a
      href={props.href}
      className="rounded-lg border border-gray-100 shadow-md p-4 flex flex-col gap-4 hover:shadow-xl hover:shadow-secondary cursor-pointer w-full"
    >
      <div className="flex flex-row justify-between items-center">
        <h2 className="text-xl font-semibold text-gray-500">{props.name}</h2>
        <div className="flex flex-row gap-2" dangerouslySetInnerHTML={{ __html: langElement }}/>
      </div>
      <p className="text-sm max-w-sm">
        {props.description}
      </p>
    </a>
  );
};
