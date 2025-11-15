import type { JSX } from "react";

export const TitleCard = (props: {
  title: string;
  children: JSX.Element;
}) => {
  return (
    <div className="flex flex-col gap-8">
      <h2 className="text-2xl">{props.title}</h2>
      {props.children}
    </div>
  );
};
