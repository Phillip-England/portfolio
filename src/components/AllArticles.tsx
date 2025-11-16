import { articleCache } from "../.."


export const AllArticles = () => {
  return (
    <ol className="flex flex-col gap-2">
      {Object.entries(articleCache.cache).map(([key, value]) => (
        <li key={key}>
          <a href={"/blog/"+value.number}>{value.number}. <span className='underline hover:text-primary'>{value.name}</span></a>
        </li>
      ))}
    </ol>
  );
};